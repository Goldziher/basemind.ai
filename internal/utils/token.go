package utils

import (
	"github.com/basemind-ai/monorepo/internal/utils/datatypes"
	"github.com/shopspring/decimal"
)

type TokenCostResult struct {
	RequestTokenCost  decimal.Decimal
	ResponseTokenCost decimal.Decimal
}

// CalculateCosts calculates the input and output tokens count and costs for a given model type / vendor.
func CalculateCosts(
	requestTokenCount int32,
	responseTokenCount int32,
	modelPricing datatypes.ProviderModelPricingDTO,
) TokenCostResult {
	// The unit size is the number of tokens per which we calculate the price. E.g. 0.002$ for 1000 tokens.
	unitSize := decimal.NewFromInt32(modelPricing.TokenUnitSize)

	return TokenCostResult{
		RequestTokenCost: decimal.NewFromInt32(requestTokenCount).
			Div(unitSize).
			Mul(modelPricing.InputTokenPrice),
		ResponseTokenCost: decimal.NewFromInt32(responseTokenCount).
			Div(unitSize).
			Mul(modelPricing.OutputTokenPrice),
	}
}
