package monk

import (
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
)

type Stance uint8

const (
	StanceNone         = 0
	FierceTiger Stance = 1 << iota
	SturdyOx
	WiseSerpent
)

const stanceEffectCategory = "Stance"

func (monk *Monk) StanceMatches(other Stance) bool {
	return (monk.Stance & other) != 0
}

func (monk *Monk) makeStanceSpell(aura *core.Aura, stanceCD *core.Timer) *core.Spell {
	return monk.RegisterSpell(core.SpellConfig{
		ActionID: aura.ActionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    stanceCD,
				Duration: time.Second,
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !aura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			aura.Activate(sim)
		},
	})
}

func (monk *Monk) registerStanceOfTheSturdyOx(stanceCD *core.Timer) {
	if monk.Spec != proto.Spec_SpecBrewmasterMonk {
		return
	}

	monk.StanceOfTheSturdyOxAura = monk.GetOrRegisterAura(core.Aura{
		Label:    "Stance of the Sturdy Ox" + monk.Label,
		ActionID: core.ActionID{SpellID: 115069},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			monk.Stance = SturdyOx
			monk.PseudoStats.DamageTakenMultiplier *= 0.75
			monk.PseudoStats.ReducedCritTakenChance += 0.06
			monk.MultiplyEnergyRegenSpeed(sim, 1.1)
			monk.SetCurrentPowerBar(core.EnergyBar)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			monk.Stance = StanceNone
			monk.PseudoStats.DamageTakenMultiplier /= 0.75
			monk.PseudoStats.ReducedCritTakenChance -= 0.06
			monk.MultiplyEnergyRegenSpeed(sim, 1.0/1.1)
		},
	})

	monk.StanceOfTheSturdyOxAura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})
	monk.StanceOfTheSturdyOx = monk.makeStanceSpell(monk.StanceOfTheSturdyOxAura, stanceCD)
}

func (monk *Monk) registerStanceOfTheWiseSerpent(stanceCD *core.Timer) {
	if monk.Spec != proto.Spec_SpecMistweaverMonk {
		return
	}

	hitDep := monk.NewDynamicStatDependency(stats.Spirit, stats.HitRating, 0.5)
	expDep := monk.NewDynamicStatDependency(stats.Spirit, stats.ExpertiseRating, 0.5)
	hasteDep := monk.NewDynamicMultiplyStat(stats.HasteRating, 1.5)
	// TODO: This should be a replacement not a dependency.
	// apDep := monk.NewDynamicStatDependency(stats.SpellPower, stats.AttackPower, 2)

	dmgDone := 0.0

	eminenceHeal := monk.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 117895},
		SpellSchool: core.SpellSchoolNature,
		ProcMask:    core.ProcMaskSpellHealing,
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

		DamageMultiplier: 0.25,
		ThreatMultiplier: 1,
		CritMultiplier:   monk.DefaultCritMultiplier(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, dmgDone, spell.OutcomeHealing)
		},
	})

	// When the Monk deals non-autoattack damage, he/she will heal the lowest health nearby target within 20 yards equal to 25% of the damage done.
	eminenceAura := monk.RegisterAura(core.Aura{
		Label:    "Eminence" + monk.Label,
		ActionID: core.ActionID{SpellID: 126890},
		Duration: core.NeverExpires,
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result == nil || !result.Landed() || result.Damage == 0 || spell.ProcMask.Matches(core.ProcMaskWhiteHit) {
				return
			}

			dmgDone = result.Damage
			// Should be a smart heal
			eminenceHeal.Cast(sim, &monk.Unit)
		},
	})

	monk.StanceOfTheWiseSerpentAura = monk.GetOrRegisterAura(core.Aura{
		Label:    "Stance of the Wise Serpent" + monk.Label,
		ActionID: core.ActionID{SpellID: 136336},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			monk.Stance = WiseSerpent
			monk.PseudoStats.HealingDealtMultiplier *= 1.2
			monk.EnableDynamicStatDep(sim, hitDep)
			monk.EnableDynamicStatDep(sim, expDep)
			monk.EnableDynamicStatDep(sim, hasteDep)
			// monk.EnableDynamicStatDep(sim, apDep)
			monk.SetCurrentPowerBar(core.ManaBar)
			eminenceAura.Activate(sim)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			eminenceAura.Deactivate(sim)
			monk.SetCurrentPowerBar(core.EnergyBar)
			// monk.DisableDynamicStatDep(sim, apDep)
			monk.DisableDynamicStatDep(sim, hasteDep)
			monk.DisableDynamicStatDep(sim, expDep)
			monk.DisableDynamicStatDep(sim, hitDep)
			monk.PseudoStats.HealingDealtMultiplier /= 1.2
			monk.Stance = StanceNone
		},
	})

	monk.StanceOfTheWiseSerpentAura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})
	monk.StanceOfTheWiseSerpent = monk.makeStanceSpell(monk.StanceOfTheWiseSerpentAura, stanceCD)
}

/*
Increases your movement speed by 10%, increases damage done by 10% and increases the amount of Chi generated by your Jab and Expel Harm abilities by 1.
*/
func (monk *Monk) registerStanceOfTheFierceTiger(stanceCD *core.Timer) {
	monk.StanceOfTheFierceTigerAura = monk.GetOrRegisterAura(core.Aura{
		Label:    "Stance of the Fierce Tiger" + monk.Label,
		ActionID: core.ActionID{SpellID: 103985},
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			monk.Stance = FierceTiger
			monk.PseudoStats.DamageDealtMultiplier *= 1.1
			// This **does** stack with other movement speed buffs.
			monk.PseudoStats.MovementSpeedMultiplier *= 1.1
			monk.SetCurrentPowerBar(core.EnergyBar)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			monk.Stance = StanceNone
			monk.PseudoStats.DamageDealtMultiplier /= 1.1
			monk.PseudoStats.MovementSpeedMultiplier /= 1.1
		},
	})

	monk.StanceOfTheFierceTigerAura.NewExclusiveEffect(stanceEffectCategory, true, core.ExclusiveEffect{})
	monk.StanceOfTheFierceTiger = monk.makeStanceSpell(monk.StanceOfTheFierceTigerAura, stanceCD)
}

func (monk *Monk) registerStances() {
	stanceCD := monk.NewTimer()
	monk.registerStanceOfTheSturdyOx(stanceCD)
	monk.registerStanceOfTheWiseSerpent(stanceCD)
	monk.registerStanceOfTheFierceTiger(stanceCD)
}
