package types

import (
	"fmt"
	"main/pkg/utils"
	"math/big"
	"strings"
)

func StateFromTendermintResponse(
	consensus *ConsensusStateResponse,
	tendermintValidators []TendermintValidator,
	chainValidators ChainValidators,
) (State, error) {
	validators, err := ValidatorsFromTendermintResponse(consensus, tendermintValidators)
	if err != nil {
		return State{}, err
	}

	hrsSplit := strings.Split(consensus.Result.RoundState.HeightRoundStep, "/")

	return State{
		Validators:      validators,
		ChainValidators: chainValidators,
		Height:          utils.MustParseInt64(hrsSplit[0]),
		Round:           utils.MustParseInt64(hrsSplit[1]),
		Step:            utils.MustParseInt64(hrsSplit[2]),
		StartTime:       consensus.Result.RoundState.StartTime,
	}, nil
}

func ValidatorsFromTendermintResponse(
	consensus *ConsensusStateResponse,
	tendermintValidators []TendermintValidator,
) (Validators, error) {
	validators := make(Validators, len(consensus.Result.RoundState.HeightVoteSet[0].Prevotes))

	for index, prevote := range consensus.Result.RoundState.HeightVoteSet[0].Prevotes {
		precommit := consensus.Result.RoundState.HeightVoteSet[0].Precommits[index]
		validator := tendermintValidators[index]

		vp := new(big.Int)
		vp, ok := vp.SetString(validator.VotingPower, 10)
		if !ok {
			return nil, fmt.Errorf("error setting string")
		}

		validators[index] = Validator{
			Address:     validator.Address,
			Precommit:   VoteFromString(precommit),
			Prevote:     VoteFromString(prevote),
			VotingPower: vp,
			IsProposer:  validator.Address == consensus.Result.RoundState.Proposer.Address,
		}
	}

	totalVP := validators.GetTotalVotingPower()

	for index, validator := range validators {
		validators[index].Index = index

		votingPowerPercent := big.NewFloat(0).SetInt(validator.VotingPower)
		votingPowerPercent = votingPowerPercent.Quo(votingPowerPercent, big.NewFloat(0).SetInt(totalVP))
		votingPowerPercent = votingPowerPercent.Mul(votingPowerPercent, big.NewFloat(100))

		validators[index].VotingPowerPercent = votingPowerPercent
	}

	return validators, nil
}

func VoteFromString(source ConsensusVote) Vote {
	if source == "nil-Vote" {
		return VotedNil
	}

	if strings.Contains(string(source), "SIGNED_MSG_TYPE_PREVOTE(Prevote) 000000000000") {
		return VotedZero
	}

	return Voted
}
