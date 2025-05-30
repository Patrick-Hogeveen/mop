package mistweaver

import (
	"time"

	"fmt"

	"github.com/wowsims/mop/sim/core"
)

func (mw *MistweaverMonk) registerRenewingMist() {
	actionID := core.ActionID{SpellID: 115151}
	chiMetrics := mw.NewChiMetrics(actionID)
	spellCoeff := 0.19665 //Will have to verify this
	targets := mw.Env.Raid.GetFirstNPlayersOrPets(int32(mw.Env.Raid.NumTargetDummies))
	maxStacks := 3

	var renewingMist *core.Spell
	renewingMist = mw.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolNature,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagHelpful | core.SpellFlagAPL,
		//ClassSpellMask: monk.MonkSpellRenewingMist,

		ManaCost: core.ManaCostOptions{BaseCostPercent: 5.85},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    mw.NewTimer(),
				Duration: time.Second * 8,
			},
		},
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   mw.DefaultCritMultiplier(),

		Hot: core.DotConfig{
			Aura: core.Aura{
				Label:     "Renewing Mist",
				MaxStacks: 3,
			},
			NumberOfTicks:        9,
			TickLength:           2 * time.Second,
			AffectedByCastSpeed:  true, //Not sure
			HasteReducesDuration: true, //Not sure
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, _ bool) {
				dot.SnapshotBaseDamage = 0 + mw.CalcScalingSpellDmg(spellCoeff)
				dot.SnapshotAttackerMultiplier = dot.Spell.CasterHealingMultiplier()
			},

			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotHealing(sim, target, dot.OutcomeTick)
				//Has to jump to two more targets after initial cast

				if maxStacks > 1 {

					for _, element := range targets {
						hot := renewingMist.Hot(element)

						if !hot.IsActive() {
							fmt.Print("Here")

							hot.Apply(sim)
							//renewingMist.Cast(sim, target)
							maxStacks = maxStacks - 1
							break
						}
					}

				}
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {

			for _, element := range targets {
				hot := spell.Hot(element)

				if !hot.IsActive() {
					hot.Apply(sim)
					break
				}
			}

			//hot := spell.Hot(target)
			//spell.Hot(spell.Unit).Apply(sim)
			//if !hot.IsActive() { //Error?
			//	hot.Apply(sim)
			//}
			chiGain := int32(1) //core.TernaryInt32(monk.StanceMatches(FierceTiger), 2, 1)
			mw.AddChi(sim, spell, chiGain, chiMetrics)

		},
	})

	mw.RegisterSpell(core.SpellConfig{})

}
