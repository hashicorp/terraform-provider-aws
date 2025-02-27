package redshiftserverless

const (
	performanceTargetLevelLowCostValue         = 1
	performanceTargetLevelEconomicalValue      = 25
	performanceTargetLevelBalancedValue        = 50
	performanceTargetLevelResourcefulValue     = 75
	performanceTargetLevelHighPerformanceValue = 100
)

const (
	performanceTargetLevelLowCost         = "LOW_COST"
	performanceTargetLevelEconomical      = "ECONOMICAL"
	performanceTargetLevelBalanced        = "BALANCED"
	performanceTargetLevelResourceful     = "RESOURCEFUL"
	performanceTargetLevelHighPerformance = "HIGH_PERFORMANCE"
)

func performanceTargetLevel_Values() []string {
	return []string{
		performanceTargetLevelLowCost,
		performanceTargetLevelEconomical,
		performanceTargetLevelBalanced,
		performanceTargetLevelResourceful,
		performanceTargetLevelHighPerformance,
	}
}
