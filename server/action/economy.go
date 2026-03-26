package action

import (
	"fmt"
	"math"

	"github.com/divinity/core/item"
	"github.com/divinity/core/knowledge"
	"github.com/divinity/core/memory"
	"github.com/divinity/core/npc"
	"github.com/divinity/core/world"
)

var economyActions = []Action{
	{
		ID: "trade", Label: "Trade goods with another NPC at the market", Category: "economy", BaseGameMinutes: 30, SkillKey: "merchant",
		Destination: destNearestOfType("market"),
		Candidates:  candidatesOfType("market"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			// Check if NPC is at ANY market, not just the first one
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.Type != "market" {
				return false
			}
			others := w.NPCsAtLocation(n.LocationID, n.ID)
			if len(others) == 0 {
				return false
			}
			hasSellable := false
			for _, it := range n.Inventory {
				if it.Name != "gold" && it.Qty > 0 {
					hasSellable = true
					break
				}
			}
			if !hasSellable {
				return false
			}
			minPrice := 9999
			for _, it := range n.Inventory {
				if it.Name != "gold" && it.Qty > 0 {
					p := w.GetPrice(it.Name)
					if p < minPrice {
						minPrice = p
					}
				}
			}
			for _, o := range others {
				if o.GoldCount() >= minPrice {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				others := w.NPCsAtLocation(n.LocationID, n.ID)
				if len(others) > 0 {
					target = others[randInt(0, len(others)-1)]
				}
			}
			if target == nil {
				return "Went to market but nobody was there."
			}
			var sellable *npc.InventoryItem
			for i := range n.Inventory {
				if n.Inventory[i].Name != "gold" && n.Inventory[i].Qty > 0 {
					sellable = &n.Inventory[i]
					break
				}
			}
			if sellable == nil {
				return "Had nothing to sell."
			}
			price := w.GetPrice(sellable.Name)
			bonus := knowledge.GetTechniqueBonus(n.ID, "trade_profit", w.Techniques)
			if bonus > 0 {
				price = int(math.Ceil(float64(price) * (1 + bonus)))
			}
			if target.GoldCount() < price {
				return fmt.Sprintf("Tried to sell %s to %s, but they couldn't afford it (price: %d gold).", sellable.Name, target.Name, price)
			}
			itemName := sellable.Name
			n.RemoveItem(itemName, 1)
			n.AddItem("gold", price)
			target.RemoveItem("gold", price)
			target.AddItem(itemName, 1)
			w.RecordTrade(itemName, true)
			n.AdjustRelationship(target.ID, 5)
			target.AdjustRelationship(n.ID, 3)
			n.GainSkill("merchant", 0.3)
			imp := 0.2
			if price > 10 {
				imp = 0.5
			}
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("Bought %s from %s for %d gold.", itemName, n.Name, price), Time: w.TimeString(), Importance: imp, Category: memory.CatEconomic, Tags: []string{n.ID, itemName}})
			return fmt.Sprintf("Sold 1 %s to %s for %d gold (market price).", itemName, target.Name, price)
		},
	},
	{
		ID: "buy_ale", Label: "Buy a drink at the inn (2 gold)", Category: "economy",
		Destination: destNearestOfType("inn"),
		Candidates:  candidatesOfType("inn"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.GoldCount() < 2 || n.HasItemOfCategory("drink") != nil {
				return false
			}
			for _, inn := range w.LocationsByType("inn") {
				barmaids := w.NPCsAtLocation(inn.ID, n.ID)
				for _, b := range barmaids {
					if b.HasProfessionOrSkill("barmaid", "innkeeping", 20) && b.HasItemOfCategory("drink") != nil {
						return true
					}
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			if n.GoldCount() < 2 {
				return "Couldn't afford a drink (need 2 gold)."
			}
			others := w.NPCsAtLocation(n.LocationID, n.ID)
			var barmaid *npc.NPC
			var drinkItem *npc.InventoryItem
			for _, b := range others {
				if b.HasProfessionOrSkill("barmaid", "innkeeping", 20) {
					if d := b.HasItemOfCategory("drink"); d != nil {
						barmaid = b
						drinkItem = d
						break
					}
				}
			}
			if barmaid == nil || drinkItem == nil {
				return "Nobody serving drinks at the inn."
			}
			drinkName := drinkItem.Name
			n.RemoveItem("gold", 2)
			barmaid.AddItem("gold", 2)
			barmaid.RemoveItem(drinkName, 1)
			n.AddItem(drinkName, 1)
			barmaid.AdjustRelationship(n.ID, 2)
			mem.Add(barmaid.ID, memory.Entry{Text: fmt.Sprintf("Sold %s to %s for 2 gold.", drinkName, n.Name), Time: w.TimeString(), Importance: 0.2, Category: memory.CatEconomic, Tags: []string{n.ID, drinkName}})
			return fmt.Sprintf("Bought %s from %s at the inn for 2 gold.", drinkName, barmaid.Name)
		},
	},
	{
		ID: "buy_food", Label: "Buy food at the market", Category: "economy",
		Destination: destNearestOfType("market"),
		Candidates:  candidatesOfType("market"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.GoldCount() < 2 {
				return false
			}
			if n.HasItemOfCategory("food") != nil {
				return false
			}
			for _, mkt := range w.LocationsByType("market") {
				sellers := w.NPCsAtLocation(mkt.ID, n.ID)
				for _, s := range sellers {
					if s.HasItemOfCategory("food") != nil {
						return true
					}
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			sellers := w.NPCsAtLocation(n.LocationID, n.ID)
			var seller *npc.NPC
			var food *npc.InventoryItem
			for _, s := range sellers {
				if f := s.HasItemOfCategory("food"); f != nil {
					seller = s
					food = f
					break
				}
			}
			if seller == nil || food == nil {
				return "Nobody at the market had food to sell."
			}
			price := w.GetPrice(food.Name)
			if n.GoldCount() < price {
				return fmt.Sprintf("Couldn't afford %s (costs %d gold).", food.Name, price)
			}
			foodName := food.Name
			n.RemoveItem("gold", price)
			seller.AddItem("gold", price)
			seller.RemoveItem(foodName, 1)
			n.AddItem(foodName, 1)
			w.RecordTrade(foodName, true)
			n.AdjustRelationship(seller.ID, 3)
			seller.AdjustRelationship(n.ID, 2)
			mem.Add(seller.ID, memory.Entry{Text: fmt.Sprintf("Sold %s to %s for %d gold.", foodName, n.Name, price), Time: w.TimeString(), Importance: 0.2, Category: memory.CatEconomic, Tags: []string{n.ID, foodName}})
			return fmt.Sprintf("Bought %s from %s for %d gold.", foodName, seller.Name, price)
		},
	},
	{
		ID: "buy_supplies", Label: "Buy crafting materials at the market", Category: "economy",
		Destination: destNearestOfType("market"),
		Candidates:  candidatesOfType("market"),
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.GoldCount() < 3 {
				return false
			}
			for _, mkt := range w.LocationsByType("market") {
				sellers := w.NPCsAtLocation(mkt.ID, n.ID)
				for _, s := range sellers {
					for _, it := range s.Inventory {
						if it.Qty > 0 {
							def := item.GetInfo(it.Name)
							if def.Category == "material" {
								return true
							}
						}
					}
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, _ *npc.NPC, w *world.World, mem memory.Store) string {
			sellers := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, seller := range sellers {
				for _, it := range seller.Inventory {
					if it.Qty > 0 {
						def := item.GetInfo(it.Name)
						if def.Category == "material" {
							price := w.GetPrice(it.Name)
							if n.GoldCount() < price {
								continue
							}
							matName := it.Name
							n.RemoveItem("gold", price)
							seller.AddItem("gold", price)
							seller.RemoveItem(matName, 1)
							n.AddItem(matName, 1)
							w.RecordTrade(matName, true)
							mem.Add(seller.ID, memory.Entry{Text: fmt.Sprintf("Sold %s to %s for %d gold.", matName, n.Name, price), Time: w.TimeString(), Importance: 0.2, Category: memory.CatEconomic, Tags: []string{n.ID, matName}})
							return fmt.Sprintf("Bought %s from %s for %d gold.", matName, seller.Name, price)
						}
					}
				}
			}
			return "Nobody at the market had supplies."
		},
	},
	{
		ID: "serve_customer", Label: "Serve food or drink to a customer at the inn", Category: "economy", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if !world.IsWorkerAtType(n, "inn", w) {
				return false
			}
			loc := w.LocationByID(n.LocationID)
			if loc == nil || loc.Type != "inn" {
				return false
			}
			hasDrink := n.HasItemOfCategory("drink") != nil
			hasFood := n.HasItemOfCategory("food") != nil
			if !hasDrink && !hasFood {
				return false
			}
			minPrice := 3
			if hasDrink {
				minPrice = 2
			}
			customers := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, c := range customers {
				if c.GoldCount() >= minPrice {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			// Prefer serving drinks first, then food
			var serving *npc.InventoryItem
			if serving = n.HasItemOfCategory("drink"); serving == nil {
				serving = n.HasItemOfCategory("food")
			}
			if serving == nil {
				return "Had nothing to serve."
			}
			def := item.GetInfo(serving.Name)
			price := 3
			if def.Category == "drink" {
				price = 2
			}
			if target == nil {
				customers := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, c := range customers {
					if c.GoldCount() >= price {
						target = c
						break
					}
				}
			}
			if target == nil {
				return "No paying customers at the inn."
			}
			if target.GoldCount() < price {
				return fmt.Sprintf("%s couldn't afford to pay %d gold.", target.Name, price)
			}
			servingName := serving.Name
			n.RemoveItem(servingName, 1)
			target.AddItem(servingName, 1)
			target.RemoveItem("gold", price)
			n.AddItem("gold", price)
			n.GainSkill("barmaid", 0.3)
			n.AdjustRelationship(target.ID, 2)
			target.AdjustRelationship(n.ID, 3)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("Bought %s from %s at the inn for %d gold.", servingName, n.Name, price), Time: w.TimeString(), Importance: 0.2, Category: memory.CatEconomic, Tags: []string{n.ID, servingName}})
			return fmt.Sprintf("Served %s to %s for %d gold.", servingName, target.Name, price)
		},
	},
	{
		ID: "heal_patient", Label: "Heal an injured or stressed NPC for gold (healer)", Category: "economy", BaseGameMinutes: 30, SkillKey: "healer",
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if !n.HasProfessionOrSkill("healer", "healer", 30) {
				return false
			}
			if n.HasItemOfCategory("medicine") == nil {
				return false
			}
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, other := range nearby {
				if (other.HP < 100 || other.Stress > 50) && other.GoldCount() >= 2 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, other := range nearby {
					if (other.HP < 100 || other.Stress > 50) && other.GoldCount() >= 2 {
						target = other
						break
					}
				}
			}
			if target == nil {
				return "No patients who can pay."
			}
			// Find best medicine
			var bestMed *npc.InventoryItem
			var bestHeal float64
			for i := range n.Inventory {
				if n.Inventory[i].Qty > 0 {
					def := item.GetInfo(n.Inventory[i].Name)
					if def.Category == "medicine" {
						h := def.Effects["heal_hp_max"]
						if h > bestHeal {
							bestMed = &n.Inventory[i]
							bestHeal = h
						}
					}
				}
			}
			if bestMed == nil {
				return "Had no healing materials."
			}
			def := item.GetInfo(bestMed.Name)
			price := 2
			if bestHeal >= 20 {
				price = 3
			}
			if target.GoldCount() < price {
				return fmt.Sprintf("%s couldn't afford healing (%d gold).", target.Name, price)
			}
			matName := bestMed.Name
			n.RemoveItem(matName, 1)
			target.RemoveItem("gold", price)
			n.AddItem("gold", price)
			hpMin := int(def.Effects["heal_hp_min"])
			hpMax := int(def.Effects["heal_hp_max"])
			if hpMin <= 0 {
				hpMin = 10
			}
			if hpMax <= hpMin {
				hpMax = hpMin + 10
			}
			hpRestore := randInt(hpMin, hpMax)
			bonus := knowledge.GetTechniqueBonus(n.ID, "healing_potency", w.Techniques)
			if bonus > 0 {
				hpRestore = int(math.Ceil(float64(hpRestore) * (1 + bonus)))
			}
			target.HP = min(100, target.HP+hpRestore)
			target.Stress = clamp(target.Stress-10, 0, 100)
			n.GainSkill("healer", 0.5)
			n.AdjustRelationship(target.ID, 5)
			target.AdjustRelationship(n.ID, 8)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("%s healed me (+%d HP) for %d gold.", n.Name, hpRestore, price), Time: w.TimeString(), Importance: 0.5, Category: memory.CatEconomic, Tags: []string{n.ID}})
			return fmt.Sprintf("Healed %s with %s (+%d HP) for %d gold.", target.Name, matName, hpRestore, price)
		},
	},
	{
		ID: "offer_counsel", Label: "Counsel a stressed NPC for gold (elder)", Category: "economy", BaseGameMinutes: 30,
		Conditions: func(n *npc.NPC, w *world.World) bool {
			if n.Profession != "elder" && n.GetEffectiveWisdom(w.GameDay, 15) < 60 {
				return false
			}
			nearby := w.NPCsAtLocation(n.LocationID, n.ID)
			for _, other := range nearby {
				if (other.Stress > 50 || other.Happiness < 30) && other.GoldCount() >= 1 {
					return true
				}
			}
			return false
		},
		Execute: func(n *npc.NPC, target *npc.NPC, w *world.World, mem memory.Store) string {
			if target == nil {
				nearby := w.NPCsAtLocation(n.LocationID, n.ID)
				for _, other := range nearby {
					if (other.Stress > 50 || other.Happiness < 30) && other.GoldCount() >= 1 {
						target = other
						break
					}
				}
			}
			if target == nil {
				return "Nobody nearby needed counsel."
			}
			if target.GoldCount() < 1 {
				return fmt.Sprintf("%s couldn't afford counsel.", target.Name)
			}
			target.RemoveItem("gold", 1)
			n.AddItem("gold", 1)
			target.Stress = clamp(target.Stress-15, 0, 100)
			target.Stats.Trauma = clamp(target.Stats.Trauma-3, 0, 100)
			target.Happiness = clamp(target.Happiness+5, 0, 100)
			n.AdjustRelationship(target.ID, 4)
			target.AdjustRelationship(n.ID, 6)
			mem.Add(target.ID, memory.Entry{Text: fmt.Sprintf("%s gave me wise counsel. I feel better.", n.Name), Time: w.TimeString(), Importance: 0.4, Category: memory.CatSocial, Tags: []string{n.ID}})
			return fmt.Sprintf("Counseled %s (-15 stress, +5 happiness) for 1 gold.", target.Name)
		},
	},
}
