package monk

import (
	"time"

	"github.com/wowsims/mop/sim/core"
)

/*
Tooltip:
Kick with a blast of Chi energy, dealing ${7.12*$<low>} to ${7.12*$<high>} Physical damage

-- Teachings of the Monastery && Stance of the Wise Serpent --

	to your target and ${3.56*$<low>} to ${3.56*$<high>} to up to 4 additional nearby targets for 50% damage

-- Teachings of the Monastery && Stance of the Wise Serpent --
.

-- Combat Conditioning --

	If behind the target, you deal an additional 20% damage over 4 sec. If in front of the target, you are instantly healed for 20% of the damage done.

-- Combat Conditioning --

-- Brewmaster Training --

	Also causes you to gain Shuffle, increasing your parry chance by 20% and your Stagger amount by an additional 20% for 6 sec.

-- Brewmaster Training --

-- Teachings of the Monastery --

	Also empowers you with Serpent's Zeal, causing you and your summoned Jade Serpent Statue to heal nearby injured targets equal to 25% of your auto-attack damage.

-- Teachings of the Monastery --
*/
var blackoutKickActionID = core.ActionID{SpellID: 100784}.WithTag(1)

func blackoutKickSpellConfig(monk *Monk, isSEFClone bool, overrides core.SpellConfig) core.SpellConfig {
	config := core.SpellConfig{
		ActionID:       blackoutKickActionID,
		SpellSchool:    core.SpellSchoolPhysical,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | SpellFlagSpender | core.SpellFlagAPL,
		ClassSpellMask: MonkSpellBlackoutKick,
		MaxRange:       core.MaxMeleeRange,

		DamageMultiplier: 7.12,
		ThreatMultiplier: 1,
		CritMultiplier:   monk.DefaultCritMultiplier(),

		Cast:               overrides.Cast,
		ExtraCastCondition: overrides.ExtraCastCondition,
		ApplyEffects:       overrides.ApplyEffects,
	}

	if isSEFClone {
		config.ActionID = config.ActionID.WithTag(SEFSpellID)
		config.Flags &= ^(core.SpellFlagAPL | SpellFlagSpender)
	}

	return config
}

func (monk *Monk) registerBlackoutKick() {
	chiMetrics := monk.NewChiMetrics(blackoutKickActionID)

	monk.RegisterSpell(blackoutKickSpellConfig(monk, false, core.SpellConfig{
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return monk.GetChi() >= 2 || monk.ComboBreakerBlackoutKickAura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := monk.CalculateMonkStrikeDamage(sim, spell)

			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if result.Landed() {
				if monk.ComboBreakerBlackoutKickAura.IsActive() {
					monk.SpendChi(sim, 0, chiMetrics)
				} else {
					monk.SpendChi(sim, 2, chiMetrics)
				}
			}

			spell.DealOutcome(sim, result)
		},
	}))
}

func (pet *StormEarthAndFirePet) registerSEFBlackoutKick() {
	pet.RegisterSpell(blackoutKickSpellConfig(pet.owner, true, core.SpellConfig{
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := pet.owner.CalculateMonkStrikeDamage(sim, spell)

			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	}))
}
