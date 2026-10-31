// Package contents keeps the items of every window, the window that is open,
// and the lists the container door needs: item names, trade offers, recipes,
// enchantment options and the creative inventory.
package contents

import (
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Item is one slot: the name, the count, and the stack as the server sent it.
type Item struct {
	Name     string
	Count    int
	Instance protocol.ItemInstance
}

// Window is the window the server opened: a chest, the bag (ContainerType
// 255), a villager's trades, a horse.
type Window struct {
	WindowID       byte
	ContainerType  byte
	Position       protocol.BlockPos
	EntityUniqueID int64
}

// TradeOffer is one offer of a villager. PriceA and PriceB are the counts
// wanted now; NetworkID is the recipe the trade is done with.
type TradeOffer struct {
	NetworkID      uint32
	BuyA, BuyB     Item
	Sell           Item
	PriceA, PriceB int
	Uses, MaxUses  int
}

// Expected is one slot an item_stack_request will change, and the item it
// will hold. The answer carries only counts and stack IDs, not the item.
type Expected struct {
	ContainerID byte
	Slot        byte
	Item        protocol.ItemInstance
}

// maxSlot is the last slot a request can name: one byte.
const maxSlot = 255

// maxWaitingRequests is how many requests may wait for the server's answer.
const maxWaitingRequests = 64

type Contents struct {
	mu             sync.Mutex
	windows        map[byte][]protocol.ItemInstance
	open           Window
	isOpen         bool
	expected       map[int32][]Expected
	nextRequestID  int32
	itemNames      map[int32]string
	itemIDs        map[string]int32
	tradeOffers    []TradeOffer
	newTradeUI     bool
	recipes        *packet.CraftingData
	enchantOptions []protocol.EnchantmentOption
	creativeItems  []protocol.CreativeItem
}

func New(items []protocol.ItemEntry) *Contents {
	c := &Contents{windows: map[byte][]protocol.ItemInstance{}, expected: map[int32][]Expected{}, nextRequestID: -1}
	c.SetItemNames(items)
	return c
}

// SetItemNames takes the item list (item_registry) that turns numbers into names.
func (c *Contents) SetItemNames(items []protocol.ItemEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// An empty list (item_registry sends one on some servers) must not wipe the list from joining.
	if len(items) == 0 && c.itemNames != nil {
		return
	}
	c.itemNames, c.itemIDs = map[int32]string{}, map[string]int32{}
	for _, entry := range items {
		c.itemNames[int32(entry.RuntimeID)] = entry.Name
		c.itemIDs[entry.Name] = int32(entry.RuntimeID)
	}
}

func (c *Contents) ItemName(networkID int32) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.itemNames[networkID]
}

func (c *Contents) ItemNetworkID(name string) (int32, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	networkID, ok := c.itemIDs[name]
	return networkID, ok
}

// Items gives the slots of a window by name. Window 0 is the player's 36
// slots (0 to 8 the hotbar), 119 the offhand, 120 the armour, 124 the UI
// (slot 0 the cursor). An open chest has the window ID it opened with.
func (c *Contents) Items(windowID byte) []Item {
	c.mu.Lock()
	defer c.mu.Unlock()
	items := make([]Item, len(c.windows[windowID]))
	for i, instance := range c.windows[windowID] {
		items[i] = Item{Count: int(instance.Stack.Count), Instance: instance}
		if !IsEmpty(instance) {
			items[i].Name = c.itemNames[instance.Stack.NetworkID]
		}
	}
	return items
}

// Cursor is what the mouse holds while a window is open.
func (c *Contents) Cursor() Item {
	instance := c.ItemIn(protocol.WindowIDUI, 0)
	if IsEmpty(instance) {
		return Item{}
	}
	return Item{Name: c.ItemName(instance.Stack.NetworkID), Count: int(instance.Stack.Count), Instance: instance}
}

// ItemIn gives one slot as the server sent it; empty when it never arrived.
func (c *Contents) ItemIn(windowID byte, slot int) protocol.ItemInstance {
	c.mu.Lock()
	defer c.mu.Unlock()
	slots := c.windows[windowID]
	if slot < 0 || slot >= len(slots) {
		return protocol.ItemInstance{}
	}
	return slots[slot]
}

func (c *Contents) SetWindowItems(windowID byte, items []protocol.ItemInstance) {
	if len(items) > maxSlot+1 {
		items = items[:maxSlot+1]
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.windows[windowID] = append([]protocol.ItemInstance(nil), items...)
}

// SetSlot keeps slots 0 to 255 only: no window of the game has more, and a
// request names a slot by one byte.
func (c *Contents) SetSlot(windowID byte, slot int, item protocol.ItemInstance) {
	if slot < 0 || slot > maxSlot {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	slots := c.windows[windowID]
	for len(slots) <= slot {
		slots = append(slots, protocol.ItemInstance{})
	}
	slots[slot] = item
	c.windows[windowID] = slots
}

// OpenWindow is the window that is open, if one is.
func (c *Contents) OpenWindow() (Window, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.open, c.isOpen
}

func (c *Contents) SetOpenWindow(window Window) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.open, c.isOpen = window, true
}

// SetClosed closes the open window and forgets its slots, unless they are the
// player's own.
// SetClosed forgets the window's items and the requests still waiting for an
// answer: they were about the window that closed.
func (c *Contents) SetClosed(windowID byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isOpen && c.open.WindowID == windowID && !isPlayerWindow(windowID) {
		delete(c.windows, windowID)
	}
	c.isOpen = false
	c.expected = map[int32][]Expected{}
}

// NextRequestID gives -1, -3, -5 ..., as a real client numbers requests.
func (c *Contents) NextRequestID() int32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextRequestID
	c.nextRequestID -= 2
	return id
}

// SetExpected remembers what a request will change. When 64 requests wait
// for an answer, the server has stopped answering: all are forgotten.
func (c *Contents) SetExpected(requestID int32, expected []Expected) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.expected) >= maxWaitingRequests {
		c.expected = map[int32][]Expected{}
	}
	c.expected[requestID] = expected
}

// TakeExpected gives what a request was expected to change and forgets it.
func (c *Contents) TakeExpected(requestID int32) ([]Expected, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	expected, ok := c.expected[requestID]
	delete(c.expected, requestID)
	return expected, ok
}

// TradeOffers are the offers of the last villager opened.
func (c *Contents) TradeOffers() []TradeOffer {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]TradeOffer(nil), c.tradeOffers...)
}

// NewTradeUI is how the server said the trade screen looks; it decides the
// containers of the ingredients.
func (c *Contents) NewTradeUI() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.newTradeUI
}

func (c *Contents) SetTrade(offers []TradeOffer, newTradeUI bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tradeOffers, c.newTradeUI = offers, newTradeUI
}

// Recipes are the recipes as crafting_data sent them.
func (c *Contents) Recipes() *packet.CraftingData {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.recipes
}

func (c *Contents) SetRecipes(recipes *packet.CraftingData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recipes = recipes
}

// RecipeOutput is what a crafting recipe makes, found by its network ID.
func (c *Contents) RecipeOutput(networkID uint32) (protocol.ItemStack, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.recipes == nil {
		return protocol.ItemStack{}, false
	}
	for _, recipe := range c.recipes.ShapedRecipes {
		if recipe.RecipeNetworkID == networkID && len(recipe.Output) > 0 {
			return recipe.Output[0], true
		}
	}
	for _, recipe := range c.recipes.ShapelessRecipes {
		if recipe.RecipeNetworkID == networkID && len(recipe.Output) > 0 {
			return recipe.Output[0], true
		}
	}
	return protocol.ItemStack{}, false
}

func (c *Contents) EnchantOptions() []protocol.EnchantmentOption {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]protocol.EnchantmentOption(nil), c.enchantOptions...)
}

func (c *Contents) SetEnchantOptions(options []protocol.EnchantmentOption) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enchantOptions = options
}

func (c *Contents) CreativeItems() []protocol.CreativeItem {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]protocol.CreativeItem(nil), c.creativeItems...)
}

func (c *Contents) SetCreativeItems(items []protocol.CreativeItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.creativeItems = items
}

// CreativeItem finds an item of the creative inventory by its item network ID,
// the one without metadata first.
func (c *Contents) CreativeItem(networkID int32) (protocol.CreativeItem, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	found, ok := protocol.CreativeItem{}, false
	for _, item := range c.creativeItems {
		if item.Item.NetworkID != networkID {
			continue
		}
		if item.Item.MetadataValue == 0 {
			return item, true
		}
		if !ok {
			found, ok = item, true
		}
	}
	return found, ok
}

// ContainerIDOf gives the container ID an item_stack_request names a slot of
// a window by. The slot numbers of the UI window (124) are the ones the game
// uses: 1-2 anvil, 3 stonecutter, 4-7 trade, 9-11 loom, 12-13 cartography,
// 14-15 enchanting, 16-17 grindstone, 27 beacon, 28-40 crafting, 50 output,
// 51-53 smithing.
func ContainerIDOf(open Window, windowID byte, slot int) (byte, bool) {
	switch windowID {
	case protocol.WindowIDInventory:
		if slot >= 0 && slot < 9 {
			return protocol.ContainerHotBar, true
		}
		return protocol.ContainerInventory, slot >= 9 && slot < 36
	case protocol.WindowIDArmour:
		return protocol.ContainerArmor, slot >= 0 && slot < 4
	case protocol.WindowIDOffHand:
		return protocol.ContainerOffhand, slot == 0
	case protocol.WindowIDUI:
		return uiContainerID(slot)
	case open.WindowID:
		switch open.ContainerType {
		case protocol.ContainerTypeFurnace, protocol.ContainerTypeBlastFurnace, protocol.ContainerTypeSmoker:
			switch {
			case slot == 0 && open.ContainerType == protocol.ContainerTypeBlastFurnace:
				return protocol.ContainerBlastFurnaceIngredient, true
			case slot == 0 && open.ContainerType == protocol.ContainerTypeSmoker:
				return protocol.ContainerSmokerIngredient, true
			case slot == 0:
				return protocol.ContainerFurnaceIngredient, true
			case slot == 1:
				return protocol.ContainerFurnaceFuel, true
			case slot == 2:
				return protocol.ContainerFurnaceResult, true
			}
			return 0, false
		case protocol.ContainerTypeBrewingStand:
			switch {
			case slot == 0:
				return protocol.ContainerBrewingStandInput, true
			case slot >= 1 && slot <= 3:
				return protocol.ContainerBrewingStandResult, true
			case slot == 4:
				return protocol.ContainerBrewingStandFuel, true
			}
			return 0, false
		}
		// Barrels and shulker boxes have their own container IDs, but their
		// window looks the same as a chest's, so they are not told apart here.
		return protocol.ContainerLevelEntity, slot >= 0 && slot <= maxSlot
	}
	return 0, false
}

func uiContainerID(slot int) (byte, bool) {
	switch {
	case slot == 0:
		return protocol.ContainerCursor, true
	case slot == 1:
		return protocol.ContainerAnvilInput, true
	case slot == 2:
		return protocol.ContainerAnvilMaterial, true
	case slot == 3:
		return protocol.ContainerStonecutterInput, true
	case slot == 4:
		return protocol.ContainerTradeTwoIngredientOne, true
	case slot == 5:
		return protocol.ContainerTradeTwoIngredientTwo, true
	case slot == 6:
		return protocol.ContainerTradeIngredientOne, true
	case slot == 7:
		return protocol.ContainerTradeIngredientTwo, true
	case slot == 9:
		return protocol.ContainerLoomInput, true
	case slot == 10:
		return protocol.ContainerLoomDye, true
	case slot == 11:
		return protocol.ContainerLoomMaterial, true
	case slot == 12:
		return protocol.ContainerCartographyInput, true
	case slot == 13:
		return protocol.ContainerCartographyAdditional, true
	case slot == 14:
		return protocol.ContainerEnchantingInput, true
	case slot == 15:
		return protocol.ContainerEnchantingMaterial, true
	case slot == 16:
		return protocol.ContainerGrindstoneInput, true
	case slot == 17:
		return protocol.ContainerGrindstoneAdditional, true
	case slot == 27:
		return protocol.ContainerBeaconPayment, true
	case slot >= 28 && slot <= 40:
		return protocol.ContainerCraftingInput, true
	case slot == 50:
		return protocol.ContainerCreatedOutput, true
	case slot == 51:
		return protocol.ContainerSmithingTableInput, true
	case slot == 52:
		return protocol.ContainerSmithingTableMaterial, true
	case slot == 53:
		return protocol.ContainerSmithingTableTemplate, true
	}
	return 0, false
}

// WindowOf gives the window and slot that a container ID and slot of an
// item_stack_response mean: the way back of ContainerIDOf.
func WindowOf(open Window, containerID, slot byte) (byte, int, bool) {
	switch containerID {
	case protocol.ContainerHotBar, protocol.ContainerInventory, protocol.ContainerCombinedHotBarAndInventory:
		return protocol.WindowIDInventory, int(slot), true
	case protocol.ContainerArmor:
		return protocol.WindowIDArmour, int(slot), true
	case protocol.ContainerOffhand:
		return protocol.WindowIDOffHand, 0, true
	case protocol.ContainerLevelEntity, protocol.ContainerBarrel, protocol.ContainerShulkerBox:
		return open.WindowID, int(slot), true
	case protocol.ContainerFurnaceIngredient, protocol.ContainerBlastFurnaceIngredient, protocol.ContainerSmokerIngredient:
		return open.WindowID, 0, true
	case protocol.ContainerFurnaceFuel:
		return open.WindowID, 1, true
	case protocol.ContainerFurnaceResult:
		return open.WindowID, 2, true
	case protocol.ContainerBrewingStandInput, protocol.ContainerBrewingStandResult, protocol.ContainerBrewingStandFuel:
		return open.WindowID, int(slot), true
	case protocol.ContainerCursor:
		return protocol.WindowIDUI, 0, true
	}
	if id, ok := uiContainerID(int(slot)); ok && id == containerID {
		return protocol.WindowIDUI, int(slot), true
	}
	return 0, 0, false
}

func IsEmpty(item protocol.ItemInstance) bool {
	return item.Stack.NetworkID == 0 || item.Stack.Count == 0
}

func isPlayerWindow(windowID byte) bool {
	switch windowID {
	case protocol.WindowIDInventory, protocol.WindowIDOffHand, protocol.WindowIDArmour, protocol.WindowIDUI:
		return true
	}
	return false
}
