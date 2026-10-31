// A villager's trade window opened, with its offers (NBT): read the offers.
package arrive

import (
	"github.com/otya0/mcbedrock-bot/internal/record/contents"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (a *Arrive) updateTrade(trade *packet.UpdateTrade) {
	var data map[string]any
	_ = nbt.Unmarshal(trade.SerialisedOffers, &data)
	recipes, _ := data["Recipes"].([]any)
	offers := make([]contents.TradeOffer, 0, len(recipes))
	for _, entry := range recipes {
		recipe, _ := entry.(map[string]any)
		offer := contents.TradeOffer{
			NetworkID: uint32(nbtInt(recipe["netId"])),
			BuyA:      a.nbtItem(recipe["buyA"]),
			BuyB:      a.nbtItem(recipe["buyB"]),
			Sell:      a.nbtItem(recipe["sell"]),
			Uses:      nbtInt(recipe["uses"]),
			MaxUses:   nbtInt(recipe["maxUses"]),
		}
		// The price is the item's count, unless buyCountA/B says a lower one (a discount).
		offer.PriceA, offer.PriceB = offer.BuyA.Count, offer.BuyB.Count
		if count := nbtInt(recipe["buyCountA"]); count > 0 {
			offer.PriceA = count
		}
		if count := nbtInt(recipe["buyCountB"]); count > 0 {
			offer.PriceB = count
		}
		// A request writes counts in one byte; an offer outside that is not usable.
		if offer.PriceA < 1 || offer.PriceA > 255 || offer.PriceB < 0 || offer.PriceB > 255 || offer.Sell.Count < 1 || offer.Sell.Count > 255 {
			continue
		}
		offers = append(offers, offer)
	}
	a.contents.SetTrade(offers, trade.NewTradeUI)
	a.contents.SetOpenWindow(contents.Window{WindowID: trade.WindowID, ContainerType: trade.WindowType, EntityUniqueID: trade.VillagerUniqueID})
}

func (a *Arrive) nbtItem(value any) contents.Item {
	compound, ok := value.(map[string]any)
	if !ok {
		return contents.Item{}
	}
	name, _ := compound["Name"].(string)
	item := contents.Item{Name: name, Count: nbtInt(compound["Count"])}
	if networkID, ok := a.contents.ItemNetworkID(name); ok {
		item.Instance.Stack.NetworkID = networkID
		item.Instance.Stack.Count = uint16(item.Count)
	}
	return item
}

func nbtInt(value any) int {
	switch v := value.(type) {
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint8:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}
