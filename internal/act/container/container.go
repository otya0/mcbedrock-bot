// Package container opens a container (the bag, a chest, a villager), works in
// it, and closes it. The work inside is one item_stack_request per call; the
// server answers with item_stack_response, which arrive writes to the record.
package container

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/otya0/mcbedrock-bot/internal/record/blocks"
	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/otya0/mcbedrock-bot/internal/record/self"
	"github.com/otya0/mcbedrock-bot/internal/serverlink"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// bagType is the bag's container type: protocol.ContainerTypeInventory is -1,
// which is 255 as a byte.
const bagType = 255

// maxStack is the most of one item a slot holds.
const maxStack = 64

// The slots of the UI window (124) that a block's window works in.
const (
	anvilInput, anvilMaterial        = 1, 2
	stonecutterInput                 = 3
	loomBanner, loomDye              = 9, 10
	enchantingInput, enchantingLapis = 14, 15
	grindstoneInput, grindstoneExtra = 16, 17
	beaconPayment                    = 27
	createdOutput                    = 50
)

// The crafting grids: the bag's 2x2 and a crafting table's 3x3, in the UI window.
var (
	bagCraftingGrid   = []int{28, 29, 30, 31}
	tableCraftingGrid = []int{32, 33, 34, 35, 36, 37, 38, 39, 40}
)

type Container struct {
	link     *serverlink.Link
	self     *self.Self
	contents *contents.Contents
	blocks   *blocks.Blocks
	logger   *slog.Logger
}

func New(link *serverlink.Link, selfRecord *self.Self, contentsRecord *contents.Contents, blocksRecord *blocks.Blocks, logger *slog.Logger) *Container {
	return &Container{link: link, self: selfRecord, contents: contentsRecord, blocks: blocksRecord, logger: logger}
}

// OpenBag presses E.
func (c *Container) OpenBag() error {
	c.logger.Info("act", "func", "Container.OpenBag", "packet", "interact/OpenInventory")
	return c.link.SendInteract(&packet.Interact{ActionType: packet.InteractActionOpenInventory, TargetEntityRuntimeID: c.self.EntityRuntimeID()})
}

// OpenOnBlock right-clicks a chest, a crafting table, a furnace and the like.
func (c *Container) OpenOnBlock(position protocol.BlockPos, face int32) error {
	if !blocks.IsFace(face) {
		return fmt.Errorf("open on block: face %d is not 0 to 5", face)
	}
	slot := c.self.HotBarSlot()
	target, _ := c.blocks.At(position)
	c.logger.Info("act", "func", "Container.OpenOnBlock", "target", blocks.PositionText(position), "face", face, "packet", "inventory_transaction/UseItem/ClickBlock")
	return c.link.SendInventoryTransaction(&packet.InventoryTransaction{TransactionData: &protocol.UseItemTransactionData{
		ActionType:       protocol.UseItemActionClickBlock,
		TriggerType:      protocol.TriggerTypePlayerInput,
		BlockPosition:    position,
		BlockFace:        face,
		HotBarSlot:       int32(slot),
		Hand:             protocol.HandSlotMainHand,
		HeldItem:         c.contents.ItemIn(protocol.WindowIDInventory, slot),
		Position:         c.self.Position(),
		ClickedPosition:  blocks.FaceCentre(face),
		BlockRuntimeID:   target.RuntimeID,
		ClientPrediction: protocol.ClientPredictionSuccess,
	}})
}

// OpenOnEntity right-clicks a villager or a horse.
func (c *Container) OpenOnEntity(runtimeID uint64) error {
	slot := c.self.HotBarSlot()
	c.logger.Info("act", "func", "Container.OpenOnEntity", "entity_runtime_id", runtimeID, "packet", "inventory_transaction/UseItemOnEntity/Interact")
	return c.link.SendInventoryTransaction(&packet.InventoryTransaction{TransactionData: &protocol.UseItemOnEntityTransactionData{
		TargetEntityRuntimeID: runtimeID,
		ActionType:            protocol.UseItemOnEntityActionInteract,
		HotBarSlot:            int32(slot),
		HeldItem:              c.contents.ItemIn(protocol.WindowIDInventory, slot),
		Position:              c.self.Position(),
		ClickedPosition:       [3]float32{0, 1, 0},
	}})
}

// TakeFrom takes the whole stack of a slot onto the cursor (a left click with
// an empty cursor). Window 0 is the bag's 36 slots, 124 the UI, or the open window.
func (c *Container) TakeFrom(window byte, slot int) error {
	open, isOpen := c.contents.OpenWindow()
	if !isOpen {
		return noWindow("take from")
	}
	from, item, err := c.slotInfo(open, window, slot)
	if err != nil {
		return err
	}
	if contents.IsEmpty(item) {
		return fmt.Errorf("take from: window %d slot %d holds nothing", window, slot)
	}
	// gophertunnel keeps Count, Source and Destination in an unnamed struct: set one by one.
	take := &protocol.TakeStackRequestAction{}
	take.Count, take.Source, take.Destination = byte(item.Stack.Count), from, c.cursorInfo()
	expected := []contents.Expected{{ContainerID: from.Container.ContainerID, Slot: from.Slot, Item: item}, {ContainerID: protocol.ContainerCursor, Slot: 0, Item: item}}
	c.logger.Info("act", "func", "Container.TakeFrom", "window", window, "slot", slot, "packet", "item_stack_request/Take")
	return c.send(c.contents.NextRequestID(), []protocol.StackRequestAction{take}, expected, nil)
}

// PlaceIn puts all the cursor holds into a slot that is empty or holds the same kind.
func (c *Container) PlaceIn(window byte, slot int) error {
	open, isOpen := c.contents.OpenWindow()
	if !isOpen {
		return noWindow("place in")
	}
	to, _, err := c.slotInfo(open, window, slot)
	if err != nil {
		return err
	}
	onCursor := c.contents.ItemIn(protocol.WindowIDUI, 0)
	if contents.IsEmpty(onCursor) {
		return errors.New("place in: the cursor holds nothing")
	}
	place := &protocol.PlaceStackRequestAction{}
	place.Count, place.Source, place.Destination = byte(onCursor.Stack.Count), c.cursorInfo(), to
	expected := []contents.Expected{{ContainerID: protocol.ContainerCursor, Slot: 0, Item: onCursor}, {ContainerID: to.Container.ContainerID, Slot: to.Slot, Item: onCursor}}
	c.logger.Info("act", "func", "Container.PlaceIn", "window", window, "slot", slot, "packet", "item_stack_request/Place")
	return c.send(c.contents.NextRequestID(), []protocol.StackRequestAction{place}, expected, nil)
}

// SwapWith swaps what the cursor holds with a slot of another kind.
func (c *Container) SwapWith(window byte, slot int) error {
	open, isOpen := c.contents.OpenWindow()
	if !isOpen {
		return noWindow("swap with")
	}
	to, inSlot, err := c.slotInfo(open, window, slot)
	if err != nil {
		return err
	}
	onCursor := c.contents.ItemIn(protocol.WindowIDUI, 0)
	swap := &protocol.SwapStackRequestAction{Source: c.cursorInfo(), Destination: to}
	expected := []contents.Expected{{ContainerID: protocol.ContainerCursor, Slot: 0, Item: inSlot}, {ContainerID: to.Container.ContainerID, Slot: to.Slot, Item: onCursor}}
	c.logger.Info("act", "func", "Container.SwapWith", "window", window, "slot", slot, "packet", "item_stack_request/Swap")
	return c.send(c.contents.NextRequestID(), []protocol.StackRequestAction{swap}, expected, nil)
}

// DestroyIn throws a slot's stack away for good (the creative inventory's bin).
func (c *Container) DestroyIn(window byte, slot int) error {
	open, isOpen := c.contents.OpenWindow()
	if !isOpen {
		return noWindow("destroy in")
	}
	from, item, err := c.slotInfo(open, window, slot)
	if err != nil {
		return err
	}
	if contents.IsEmpty(item) {
		return fmt.Errorf("destroy in: window %d slot %d holds nothing", window, slot)
	}
	destroy := &protocol.DestroyStackRequestAction{Count: byte(item.Stack.Count), Source: from}
	expected := []contents.Expected{{ContainerID: from.Container.ContainerID, Slot: from.Slot, Item: item}}
	c.logger.Info("act", "func", "Container.DestroyIn", "window", window, "slot", slot, "packet", "item_stack_request/Destroy")
	return c.send(c.contents.NextRequestID(), []protocol.StackRequestAction{destroy}, expected, nil)
}

// TakeFromCreative takes count of an item ("minecraft:stone") from the creative
// inventory into a slot. The created stack's ID is the request ID, which the
// server accepted in the version 2 experiments.
func (c *Container) TakeFromCreative(item string, count int, window byte, slot int) error {
	open, isOpen := c.contents.OpenWindow()
	if !isOpen {
		return noWindow("take from creative")
	}
	if count < 1 || count > maxStack {
		return fmt.Errorf("take from creative: count %d is not 1 to %d", count, maxStack)
	}
	networkID, ok := c.contents.ItemNetworkID(item)
	if !ok {
		return fmt.Errorf("take from creative: no item is called %q", item)
	}
	creative, ok := c.contents.CreativeItem(networkID)
	if !ok {
		return fmt.Errorf("take from creative: %q is not in the creative inventory", item)
	}
	to, _, err := c.slotInfo(open, window, slot)
	if err != nil {
		return err
	}
	requestID := c.contents.NextRequestID()
	taken := protocol.ItemInstance{Stack: creative.Item}
	taken.Stack.Count = uint16(count)
	place, placed := c.output(requestID, count, to, taken)
	actions := []protocol.StackRequestAction{&protocol.CraftCreativeStackRequestAction{CreativeItemNetworkID: creative.CreativeItemNetworkID, NumberOfCrafts: 1}, place}
	c.logger.Info("act", "func", "Container.TakeFromCreative", "item", item, "count", count, "window", window, "slot", slot, "packet", "item_stack_request/CraftCreative,Place")
	return c.send(requestID, actions, []contents.Expected{placed}, nil)
}

// Craft makes a recipe (its network ID, from crafting_data) out of the items in
// the crafting grid of the bag or a crafting table, or the stonecutter's input,
// and puts the result into a slot.
func (c *Container) Craft(recipeNetworkID uint32, window byte, slot int) error {
	open, isOpen := c.contents.OpenWindow()
	if !isOpen {
		return noWindow("craft")
	}
	var inputs []int
	switch open.ContainerType {
	case bagType:
		inputs = bagCraftingGrid
	case protocol.ContainerTypeWorkbench:
		inputs = tableCraftingGrid
	case protocol.ContainerTypeStonecutter:
		inputs = []int{stonecutterInput}
	default:
		return fmt.Errorf("craft: window type %d is not a crafting window", open.ContainerType)
	}
	output, ok := c.contents.RecipeOutput(recipeNetworkID)
	if !ok {
		return fmt.Errorf("craft: no crafting recipe has network ID %d", recipeNetworkID)
	}
	to, _, err := c.slotInfo(open, window, slot)
	if err != nil {
		return err
	}
	requestID := c.contents.NextRequestID()
	consumes, expected := c.consumeUI(inputs, 1)
	to = emptied(to, expected, 1)
	place, placed := c.output(requestID, int(output.Count), to, protocol.ItemInstance{Stack: output})
	actions := []protocol.StackRequestAction{&protocol.CraftRecipeStackRequestAction{RecipeNetworkID: recipeNetworkID, NumberOfCrafts: 1}}
	actions = append(actions, consumes...)
	actions = append(actions, place)
	c.logger.Info("act", "func", "Container.Craft", "recipe_network_id", recipeNetworkID, "window", window, "slot", slot, "packet", "item_stack_request/CraftRecipe,Consume,Place")
	return c.send(requestID, actions, append(expected, placed), nil)
}

// Trade does one offer of the villager's window: what it wants moves from bag
// slots slotA and slotB (slotB only when it wants two things) into the trade
// slots, the trade is made, and what it gives goes into bag slot toSlot.
func (c *Container) Trade(offer int, slotA, slotB, toSlot int) error {
	open, err := c.openOfType("trade", protocol.ContainerTypeTrade)
	if err != nil {
		return err
	}
	offers := c.contents.TradeOffers()
	if offer < 0 || offer >= len(offers) {
		return fmt.Errorf("trade: offer %d is not one of the %d offers", offer, len(offers))
	}
	chosen := offers[offer]
	// The new trade screen and the old one name the two ingredient slots differently.
	containerA, tradeSlotA := byte(protocol.ContainerTradeTwoIngredientOne), byte(4)
	containerB, tradeSlotB := byte(protocol.ContainerTradeTwoIngredientTwo), byte(5)
	if !c.contents.NewTradeUI() {
		containerA, tradeSlotA = protocol.ContainerTradeIngredientOne, 6
		containerB, tradeSlotB = protocol.ContainerTradeIngredientTwo, 7
	}
	to, _, err := c.slotInfo(open, protocol.WindowIDInventory, toSlot)
	if err != nil {
		return err
	}
	requestID := c.contents.NextRequestID()
	placeA, consumeA, takenA, err := c.moveIntoTrade(open, slotA, chosen.PriceA, containerA, tradeSlotA, requestID)
	if err != nil {
		return err
	}
	places := []protocol.StackRequestAction{placeA}
	consumes := []protocol.StackRequestAction{consumeA}
	expected := []contents.Expected{takenA}
	to = emptied(to, expected, chosen.PriceA)
	if chosen.BuyB.Name != "" {
		placeB, consumeB, takenB, err := c.moveIntoTrade(open, slotB, chosen.PriceB, containerB, tradeSlotB, requestID)
		if err != nil {
			return err
		}
		places = append(places, placeB)
		consumes = append(consumes, consumeB)
		expected = append(expected, takenB)
		to = emptied(to, []contents.Expected{takenB}, chosen.PriceB)
	}
	place, placed := c.output(requestID, chosen.Sell.Count, to, chosen.Sell.Instance)
	actions := append(places, &protocol.CraftRecipeStackRequestAction{RecipeNetworkID: chosen.NetworkID, NumberOfCrafts: 1})
	actions = append(actions, consumes...)
	actions = append(actions, place)
	c.logger.Info("act", "func", "Container.Trade", "offer", offer, "slot_a", slotA, "slot_b", slotB, "to_slot", toSlot, "packet", "item_stack_request/Place,CraftRecipe,Consume,Place")
	return c.send(requestID, actions, append(expected, placed), nil)
}

// moveIntoTrade moves count of a bag slot into a trade slot (Place) and uses
// it up there (Consume). A part of a stack becomes a new stack whose ID is the
// request ID; a whole stack keeps its ID.
func (c *Container) moveIntoTrade(open contents.Window, fromSlot, count int, containerID, tradeSlot byte, requestID int32) (place, consume protocol.StackRequestAction, taken contents.Expected, err error) {
	from, item, err := c.slotInfo(open, protocol.WindowIDInventory, fromSlot)
	if err != nil {
		return nil, nil, taken, err
	}
	if contents.IsEmpty(item) || int(item.Stack.Count) < count {
		return nil, nil, taken, fmt.Errorf("trade: bag slot %d holds %d, the offer wants %d", fromSlot, item.Stack.Count, count)
	}
	tradeSlotInfo := protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: containerID}, Slot: tradeSlot}
	moved := &protocol.PlaceStackRequestAction{}
	moved.Count, moved.Source, moved.Destination = byte(count), from, tradeSlotInfo
	used := &protocol.ConsumeStackRequestAction{}
	used.Count, used.Source = byte(count), tradeSlotInfo
	used.Source.StackNetworkID = item.StackNetworkID
	if int(item.Stack.Count) > count {
		used.Source.StackNetworkID = requestID
	}
	taken = contents.Expected{ContainerID: from.Container.ContainerID, Slot: from.Slot, Item: item}
	return moved, used, taken, nil
}

// Enchant picks one of the enchanting table's three options (0 to 2): the item
// and the lapis are used up, and the enchanted item goes back into the input.
func (c *Container) Enchant(option int) error {
	open, err := c.openOfType("enchant", protocol.ContainerTypeEnchantment)
	if err != nil {
		return err
	}
	options := c.contents.EnchantOptions()
	if option < 0 || option >= len(options) {
		return fmt.Errorf("enchant: option %d is not one of the %d options", option, len(options))
	}
	input := c.contents.ItemIn(protocol.WindowIDUI, enchantingInput)
	to, _, err := c.slotInfo(open, protocol.WindowIDUI, enchantingInput)
	if err != nil {
		return err
	}
	requestID := c.contents.NextRequestID()
	consumeItem, expectedItem := c.consumeUI([]int{enchantingInput}, 1)
	// Option n costs n+1 lapis.
	consumeLapis, expectedLapis := c.consumeUI([]int{enchantingLapis}, option+1)
	to = emptied(to, expectedItem, 1)
	result := input
	result.Stack.Count = 1
	place, placed := c.output(requestID, 1, to, result)
	actions := []protocol.StackRequestAction{&protocol.CraftRecipeStackRequestAction{RecipeNetworkID: options[option].RecipeNetworkID, NumberOfCrafts: 1}}
	actions = append(actions, consumeItem...)
	actions = append(actions, consumeLapis...)
	actions = append(actions, place)
	expected := append(expectedItem, expectedLapis...)
	c.logger.Info("act", "func", "Container.Enchant", "option", option, "packet", "item_stack_request/CraftRecipe,Consume,Place")
	return c.send(requestID, actions, append(expected, placed), nil)
}

// Rename names the item in the anvil (with the material in the second slot, if
// any) and puts the result into a slot.
func (c *Container) Rename(name string, window byte, slot int) error {
	open, err := c.openOfType("rename", protocol.ContainerTypeAnvil)
	if err != nil {
		return err
	}
	c.logger.Info("act", "func", "Container.Rename", "name", name, "window", window, "slot", slot, "packet", "item_stack_request/CraftRecipeOptional,Consume,Place")
	work := &protocol.CraftRecipeOptionalStackRequestAction{FilterStringIndex: 0}
	return c.craftInUI(open, []int{anvilInput, anvilMaterial}, window, slot, work, []string{name})
}

// Grind takes the enchantments off with the grindstone.
func (c *Container) Grind(window byte, slot int) error {
	open, err := c.openOfType("grind", protocol.ContainerTypeGrindstone)
	if err != nil {
		return err
	}
	c.logger.Info("act", "func", "Container.Grind", "window", window, "slot", slot, "packet", "item_stack_request/CraftGrindstoneRecipe,Consume,Place")
	work := &protocol.CraftGrindstoneRecipeStackRequestAction{NumberOfCrafts: 1}
	return c.craftInUI(open, []int{grindstoneInput, grindstoneExtra}, window, slot, work, nil)
}

// Weave puts a pattern ("bo", "cr" ...) on the banner in the loom with its dye.
func (c *Container) Weave(pattern string, window byte, slot int) error {
	open, err := c.openOfType("weave", protocol.ContainerTypeLoom)
	if err != nil {
		return err
	}
	c.logger.Info("act", "func", "Container.Weave", "pattern", pattern, "window", window, "slot", slot, "packet", "item_stack_request/CraftLoomRecipe,Consume,Place")
	work := &protocol.CraftLoomRecipeStackRequestAction{Pattern: pattern, TimesCrafted: 1}
	return c.craftInUI(open, []int{loomBanner, loomDye}, window, slot, work, nil)
}

// PayBeacon pays the beacon with the item in its slot and picks its effects.
func (c *Container) PayBeacon(primary, secondary int32) error {
	_, err := c.openOfType("pay beacon", protocol.ContainerTypeBeacon)
	if err != nil {
		return err
	}
	consumes, expected := c.consumeUI([]int{beaconPayment}, 1)
	actions := []protocol.StackRequestAction{&protocol.BeaconPaymentStackRequestAction{PrimaryEffect: primary, SecondaryEffect: secondary}}
	actions = append(actions, consumes...)
	c.logger.Info("act", "func", "Container.PayBeacon", "primary", primary, "secondary", secondary, "packet", "item_stack_request/BeaconPayment,Consume")
	return c.send(c.contents.NextRequestID(), actions, expected, nil)
}

// Close closes the open window. What the cursor holds is not put back: do that
// first with PlaceIn, as a player does.
func (c *Container) Close() error {
	open, isOpen := c.contents.OpenWindow()
	if !isOpen {
		return noWindow("close")
	}
	c.logger.Info("act", "func", "Container.Close", "window", open.WindowID, "packet", "container_close")
	return c.link.SendContainerClose(&packet.ContainerClose{WindowID: open.WindowID, ContainerType: open.ContainerType})
}

// craftInUI is the shape of the anvil, the grindstone and the loom: the work,
// each input used up by one, and the result into a slot.
func (c *Container) craftInUI(open contents.Window, inputs []int, window byte, slot int, work protocol.StackRequestAction, filterStrings []string) error {
	to, _, err := c.slotInfo(open, window, slot)
	if err != nil {
		return err
	}
	result := c.contents.ItemIn(protocol.WindowIDUI, inputs[0])
	result.Stack.Count = 1
	requestID := c.contents.NextRequestID()
	consumes, expected := c.consumeUI(inputs, 1)
	to = emptied(to, expected, 1)
	place, placed := c.output(requestID, 1, to, result)
	actions := []protocol.StackRequestAction{work}
	actions = append(actions, consumes...)
	actions = append(actions, place)
	return c.send(requestID, actions, append(expected, placed), filterStrings)
}

// consumeUI uses up count of each UI slot that holds something.
func (c *Container) consumeUI(slots []int, count int) ([]protocol.StackRequestAction, []contents.Expected) {
	var actions []protocol.StackRequestAction
	var expected []contents.Expected
	for _, slot := range slots {
		item := c.contents.ItemIn(protocol.WindowIDUI, slot)
		containerID, ok := contents.ContainerIDOf(contents.Window{}, protocol.WindowIDUI, slot)
		if contents.IsEmpty(item) || !ok {
			continue
		}
		consume := &protocol.ConsumeStackRequestAction{}
		consume.Count = byte(count)
		consume.Source = protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: containerID}, Slot: byte(slot), StackNetworkID: item.StackNetworkID}
		actions = append(actions, consume)
		expected = append(expected, contents.Expected{ContainerID: containerID, Slot: byte(slot), Item: item})
	}
	return actions, expected
}

// output moves count of what was made, from the created output (whose stack ID
// is the request ID) into a slot.
func (c *Container) output(requestID int32, count int, to protocol.StackRequestSlotInfo, made protocol.ItemInstance) (protocol.StackRequestAction, contents.Expected) {
	place := &protocol.PlaceStackRequestAction{}
	place.Count = byte(count)
	place.Source = protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: protocol.ContainerCreatedOutput}, Slot: createdOutput, StackNetworkID: requestID}
	place.Destination = to
	return place, contents.Expected{ContainerID: to.Container.ContainerID, Slot: to.Slot, Item: made}
}

// emptied gives the destination the stack ID 0 when an earlier action of the
// same request takes count out of that slot and leaves it empty.
func emptied(to protocol.StackRequestSlotInfo, taken []contents.Expected, count int) protocol.StackRequestSlotInfo {
	for _, from := range taken {
		sameSlot := from.ContainerID == to.Container.ContainerID && from.Slot == to.Slot
		if sameSlot && int(from.Item.Stack.Count) <= count {
			to.StackNetworkID = 0
		}
	}
	return to
}

// openOfType gives the open window when it is of that type.
func (c *Container) openOfType(function string, containerType byte) (contents.Window, error) {
	open, isOpen := c.contents.OpenWindow()
	if !isOpen || open.ContainerType != containerType {
		return contents.Window{}, fmt.Errorf("%s: no window of type %d is open", function, containerType)
	}
	return open, nil
}

func noWindow(function string) error {
	return errors.New(function + ": no window is open")
}

// slotInfo names a slot of a window the way a request does, with what it holds.
func (c *Container) slotInfo(open contents.Window, window byte, slot int) (protocol.StackRequestSlotInfo, protocol.ItemInstance, error) {
	containerID, ok := contents.ContainerIDOf(open, window, slot)
	if !ok {
		return protocol.StackRequestSlotInfo{}, protocol.ItemInstance{}, fmt.Errorf("window %d slot %d is not a slot of the open window", window, slot)
	}
	item := c.contents.ItemIn(window, slot)
	return protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: containerID}, Slot: byte(slot), StackNetworkID: item.StackNetworkID}, item, nil
}

func (c *Container) cursorInfo() protocol.StackRequestSlotInfo {
	onCursor := c.contents.ItemIn(protocol.WindowIDUI, 0)
	return protocol.StackRequestSlotInfo{Container: protocol.FullContainerName{ContainerID: protocol.ContainerCursor}, StackNetworkID: onCursor.StackNetworkID}
}

// send sends one item_stack_request and remembers what it should change.
func (c *Container) send(requestID int32, actions []protocol.StackRequestAction, expected []contents.Expected, filterStrings []string) error {
	c.contents.SetExpected(requestID, expected)
	return c.link.SendItemStackRequest(&packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{RequestID: requestID, Actions: actions, FilterStrings: filterStrings}}})
}
