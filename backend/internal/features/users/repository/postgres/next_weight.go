package users_postgres_repository

const (
	progressionStepFraction          = 0.10
	progressionDeviationThreshold    = 0.25
	progressionConsolidationFraction = 0.10
	normRepsMin                      = 6
	normRepsMax                      = 12
)

func NextRecommendedWeight(rec, act float64, r int) float64 {
	if rec < 0 {
		rec = 0
	}
	if act < 0 {
		act = 0
	}
	if rec <= 0 {
		rec = act
	}
	if act <= 0 {
		act = rec
	}

	step := progressionStepFraction * rec
	if step <= 0 {
		step = gymPlateStepSmallKG
	}

	consolidationDrop := progressionConsolidationFraction * act
	minFloor := gymPlateStepSmallKG

	var next float64
	switch {
	case r < normRepsMin:
		next = max(act-step, minFloor)
	case r >= normRepsMax:
		next = act + step
	default:
		if act >= rec*(1.0+progressionDeviationThreshold) {
			next = act - consolidationDrop
		} else if act < rec {
			next = max(act, rec-step)
		} else {
			next = act
		}
	}

	if next < minFloor {
		next = minFloor
	}
	return RoundWeightToNearestGymStep(next)
}