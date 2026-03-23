package world

import "math"

type EconomyEntry struct {
	Supply float64 `json:"supply"`
	Demand float64 `json:"demand"`
}

func (w *World) GetPrice(itemName string) int {
	base := 2
	if p, ok := w.BasePrices[itemName]; ok {
		base = p
	}
	eco, ok := w.Economy[itemName]
	if !ok {
		return base
	}
	mult := 1 + (eco.Demand-eco.Supply)/100
	clamped := math.Max(w.MinMultiplier, math.Min(w.MaxMultiplier, mult))
	return max(1, int(math.Round(float64(base)*clamped)))
}

func (w *World) RecordTrade(itemName string, isSale bool) {
	if _, ok := w.Economy[itemName]; !ok {
		w.Economy[itemName] = EconomyEntry{Supply: 50, Demand: 50}
	}
	e := w.Economy[itemName]
	if isSale {
		e.Supply = math.Max(0, e.Supply-2)
		e.Demand = math.Min(100, e.Demand+1)
	} else {
		e.Supply = math.Min(100, e.Supply+2)
		e.Demand = math.Max(0, e.Demand-1)
	}
	w.Economy[itemName] = e
}

func (w *World) DecayEconomy() {
	for k, e := range w.Economy {
		e.Supply += (50 - e.Supply) * 0.05
		e.Demand += (50 - e.Demand) * 0.05
		w.Economy[k] = e
	}
}
