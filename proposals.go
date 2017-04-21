// Copyright (c) 2017 Dave Collins
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/davecgh/dcrstakesim/internal/tickettreap"
	"github.com/decred/dcrutil"
	"math"
)

// calcNextStakeDiffProposal1 returns the required stake difficulty (aka ticket
// price) for the block after the current tip block the simulator is associated
// with using the algorithm proposed by raedah in
// https://github.com/decred/dcrd/issues/584
func (s *simulator) calcNextStakeDiffProposal1() int64 {
	// Stake difficulty before any tickets could possibly be purchased is
	// the minimum value.
	nextHeight := int32(0)
	if s.tip != nil {
		nextHeight = s.tip.height + 1
	}
	stakeDiffStartHeight := int32(s.params.CoinbaseMaturity) + 1
	if nextHeight < stakeDiffStartHeight {
		return s.params.MinimumStakeDiff
	}

	// Return the previous block's difficulty requirements if the next block
	// is not at a difficulty retarget interval.
	intervalSize := s.params.StakeDiffWindowSize
	curDiff := s.tip.ticketPrice
	if int64(nextHeight)%intervalSize != 0 {
		return curDiff
	}

	// Attempt to get the pool size from the previous retarget interval.
	var prevPoolSize int64
	prevRetargetHeight := nextHeight - int32(intervalSize)
	node := s.ancestorNode(s.tip, prevRetargetHeight, nil)
	if node != nil {
		prevPoolSize = int64(node.poolSize)
	}

	// Return the existing ticket price for the first interval.
	if prevPoolSize == 0 {
		return curDiff
	}

	curPoolSize := int64(s.tip.poolSize)
	ratio := float64(curPoolSize) / float64(prevPoolSize)
	return int64(float64(curDiff) * ratio)
}

// the algorithm proposed by raedah (enhanced older)
func (s *simulator) calcNextStakeDiffProposal1E() int64 {
	// Stake difficulty before any tickets could possibly be purchased is
	// the minimum value.
	nextHeight := int32(0)
	if s.tip != nil {
		nextHeight = s.tip.height + 1
	}
	stakeDiffStartHeight := int32(s.params.CoinbaseMaturity) + 1
	if nextHeight < stakeDiffStartHeight {
		return s.params.MinimumStakeDiff
	}

	// Return the previous block's difficulty requirements if the next block
	// is not at a difficulty retarget interval.
	intervalSize := s.params.StakeDiffWindowSize
	curDiff := s.tip.ticketPrice
	if int64(nextHeight)%intervalSize != 0 {
		return curDiff
	}

	// Attempt to get the pool size from the previous retarget interval.
	var prevPoolSize int64
	prevRetargetHeight := nextHeight - int32(intervalSize)
	node := s.ancestorNode(s.tip, prevRetargetHeight, nil)
	if node != nil {
		prevPoolSize = int64(node.poolSize)
	}

	// Return the existing ticket price for the first interval.
	if prevPoolSize == 0 {
		return curDiff
	}

	// get the immature ticket count from the previous window
	// note, make sure we have no off-by-ones here
	var prevImmatureTickets int64
	ticketMaturity := int64(s.params.TicketMaturity)
	relevantHeight := s.tip.height - int32(intervalSize) // or nextHeight?
	relevantNode := s.ancestorNode(s.tip, relevantHeight, nil)
	s.ancestorNode(relevantNode, relevantHeight-int32(ticketMaturity), func(n *blockNode) {
		prevImmatureTickets += int64(len(n.ticketsAdded))
	})

	// derive ratio of percent change in pool size
	// max possible poolSizeChangeRatio is 2
	immatureTickets := int64(len(s.immatureTickets))
	curPoolSize := int64(s.tip.poolSize)
	curPoolSizeAll := curPoolSize + immatureTickets
	prevPoolSizeAll := prevPoolSize + prevImmatureTickets
	poolSizeChangeRatio := float64(curPoolSizeAll) / float64(prevPoolSizeAll)

	// derive ratio of percent of target pool size
	ticketsPerBlock := int64(s.params.TicketsPerBlock)
	ticketPoolSize := int64(s.params.TicketPoolSize)
	targetPoolSize := ticketsPerBlock * ticketPoolSize
	targetPoolSizeAll := ticketsPerBlock * (ticketPoolSize + ticketMaturity)
	targetRatio := float64(curPoolSizeAll) / float64(targetPoolSizeAll)

	// Voila!
	nextDiff := float64(curDiff) * poolSizeChangeRatio * targetRatio

	// insure the pool gets fully populated
	maximumStakeDiff := int64(float64(s.tip.totalSupply) / float64(targetPoolSize))
	if int64(nextDiff) > maximumStakeDiff {
		if maximumStakeDiff < s.params.MinimumStakeDiff {
			return s.params.MinimumStakeDiff
		}
		return maximumStakeDiff
	}

	// hard coded minimum value
	if int64(nextDiff) < s.params.MinimumStakeDiff {
		return s.params.MinimumStakeDiff
	}

	return int64(nextDiff)
}

// the algorithm proposed by raedah (enhanced newer F)
func (s *simulator) calcNextStakeDiffProposal1F() int64 {
	// Stake difficulty before any tickets could possibly be purchased is
	// the minimum value.
	nextHeight := int32(0)
	if s.tip != nil {
		nextHeight = s.tip.height + 1
	}
	stakeDiffStartHeight := int32(s.params.CoinbaseMaturity) + 1
	if nextHeight < stakeDiffStartHeight {
		return s.params.MinimumStakeDiff
	}

	// Return the previous block's difficulty requirements if the next block
	// is not at a difficulty retarget interval.
	intervalSize := s.params.StakeDiffWindowSize
	curDiff := s.tip.ticketPrice
	if int64(nextHeight)%intervalSize != 0 {
		return curDiff
	}

	// Attempt to get the pool size from the previous retarget interval.
	var prevPoolSize int64
	prevRetargetHeight := nextHeight - int32(intervalSize)
	node := s.ancestorNode(s.tip, prevRetargetHeight, nil)
	if node != nil {
		prevPoolSize = int64(node.poolSize)
	}

	// Return the existing ticket price for the first interval.
	if prevPoolSize == 0 {
		return curDiff
	}

	// get the immature ticket count from the previous window
	var prevImmatureTickets int64
	ticketMaturity := int64(s.params.TicketMaturity)
	relevantHeight := s.tip.height - int32(intervalSize)
	relevantNode := s.ancestorNode(s.tip, relevantHeight, nil)
	s.ancestorNode(relevantNode, relevantHeight-int32(ticketMaturity), func(n *blockNode) {
		prevImmatureTickets += int64(len(n.ticketsAdded))
	})

	// derive ratio of percent change in pool size
	// max possible poolSizeChangeRatio is 2
	immatureTickets := int64(len(s.immatureTickets))
	curPoolSize := int64(s.tip.poolSize)
	curPoolSizeAll := curPoolSize + immatureTickets
	prevPoolSizeAll := prevPoolSize + prevImmatureTickets
	poolSizeChangeRatio := float64(curPoolSizeAll) / float64(prevPoolSizeAll)

	// derive ratio of percent of target pool size
	ticketsPerBlock := int64(s.params.TicketsPerBlock)
	ticketsPerWindow := ticketsPerBlock * intervalSize
	ticketPoolSize := int64(s.params.TicketPoolSize)
	targetPoolSize := ticketsPerBlock * ticketPoolSize
	targetPoolSizeAll := ticketsPerBlock * (ticketPoolSize + ticketMaturity)
	targetRatio := float64(curPoolSizeAll) / float64(targetPoolSizeAll)

	// Voila!
	var nextDiff float64
	if poolSizeChangeRatio < 1.0 {
		// Upward price movements are stronger then downward movements.
		// Add downward movements relative strength, for the market to respond and give its input.
		maxFreshStakePerBlock := int64(s.params.MaxFreshStakePerBlock)
		maxFreshStakePerWindow := maxFreshStakePerBlock * intervalSize
		buysPerVote := float64(maxFreshStakePerWindow) / float64(ticketsPerWindow)
		sizeDiff := float64(prevPoolSizeAll) - float64(curPoolSizeAll)
		tempPoolSizeChangeRatio := (float64(prevPoolSizeAll) - (sizeDiff * buysPerVote)) / float64(prevPoolSizeAll)
		nextDiff = float64(curDiff) * tempPoolSizeChangeRatio * targetRatio
	} else {
		// strength of gravity for acceleration above target pool size
		relativeIntervals := math.Abs(float64(targetPoolSizeAll-curPoolSizeAll)) / float64(ticketsPerWindow)
		nextDiff = float64(curDiff) * math.Pow(poolSizeChangeRatio, relativeIntervals) * targetRatio
	}

	// ramp up price during initial pool population
	maximumStakeDiff := int64(float64(s.tip.totalSupply) / float64(targetPoolSize))
	if int64(nextDiff) > maximumStakeDiff && targetRatio < 1.0 {
		nextDiff = float64(maximumStakeDiff) * targetRatio
	}

	// optional
	if int64(nextDiff) > maximumStakeDiff {
		if maximumStakeDiff < s.params.MinimumStakeDiff {
			return s.params.MinimumStakeDiff
		}
		return maximumStakeDiff
	}

	// hard coded minimum value
	if int64(nextDiff) < s.params.MinimumStakeDiff {
		return s.params.MinimumStakeDiff
	}

	return int64(nextDiff)
}

// the algorithm proposed by raedah (enhanced newer)
func (s *simulator) calcNextStakeDiffProposal1G() int64 {
	// Stake difficulty before any tickets could possibly be purchased is
	// the minimum value.
	nextHeight := int32(0)
	if s.tip != nil {
		nextHeight = s.tip.height + 1
	}
	stakeDiffStartHeight := int32(s.params.CoinbaseMaturity) + 1
	if nextHeight < stakeDiffStartHeight {
		return s.params.MinimumStakeDiff
	}

	// Return the previous block's difficulty requirements if the next block
	// is not at a difficulty retarget interval.
	intervalSize := s.params.StakeDiffWindowSize
	curDiff := s.tip.ticketPrice
	if int64(nextHeight)%intervalSize != 0 {
		return curDiff
	}

	// Attempt to get the pool size from the previous retarget interval.
	var prevPoolSize int64
	prevRetargetHeight := nextHeight - int32(intervalSize)
	node := s.ancestorNode(s.tip, prevRetargetHeight, nil)
	if node != nil {
		prevPoolSize = int64(node.poolSize)
	}

	// Return the existing ticket price for the first interval.
	if prevPoolSize == 0 {
		return curDiff
	}

	// get the immature ticket count from the previous window
	// note, make sure we have no off-by-ones here
	var prevImmatureTickets int64
	ticketMaturity := int64(s.params.TicketMaturity)
	relevantHeight := s.tip.height - int32(intervalSize) // or nextHeight?
	relevantNode := s.ancestorNode(s.tip, relevantHeight, nil)
	s.ancestorNode(relevantNode, relevantHeight-int32(ticketMaturity), func(n *blockNode) {
		prevImmatureTickets += int64(len(n.ticketsAdded))
	})

	// derive ratio of percent change in pool size
	// max possible poolSizeChangeRatio is 2
	immatureTickets := int64(len(s.immatureTickets))
	curPoolSize := int64(s.tip.poolSize)
	curPoolSizeAll := curPoolSize + immatureTickets
	prevPoolSizeAll := prevPoolSize + prevImmatureTickets
	poolSizeChangeRatio := float64(curPoolSizeAll) / float64(prevPoolSizeAll)

	// derive ratio of percent of target pool size
	ticketsPerBlock := int64(s.params.TicketsPerBlock)
	ticketsPerWindow := ticketsPerBlock * intervalSize
	ticketPoolSize := int64(s.params.TicketPoolSize)
	targetPoolSize := ticketsPerBlock * ticketPoolSize
	targetPoolSizeAll := ticketsPerBlock * (ticketPoolSize + ticketMaturity)
	targetRatio := float64(curPoolSizeAll) / float64(targetPoolSizeAll)

	// derive ratio of purchase slots filled
	maxFreshStakePerBlock := int64(s.params.MaxFreshStakePerBlock)
	maxFreshStakePerWindow := maxFreshStakePerBlock * intervalSize
	freshStakeLastWindow := curPoolSizeAll - prevPoolSizeAll
	// steady is a consistent flow of tickets in and out
	// mainnet steady is <0.25 is a drop, >0.25 is a rise
	steadyFreshStakeRatio := float64(ticketsPerBlock) / float64(maxFreshStakePerBlock)
	freshStakeRatio := (float64(freshStakeLastWindow) / float64(maxFreshStakePerWindow)) * (1.0 / steadyFreshStakeRatio)

	// Upward price movements are stronger then downward movements.
	// Add downward movements relative strength, for the market to respond and give its input.
	if poolSizeChangeRatio < 1.0 {
		buysPerVote := float64(maxFreshStakePerWindow) / float64(ticketsPerWindow)
		sizeDiff := float64(prevPoolSizeAll - curPoolSizeAll)
		poolSizeChangeRatio = (float64(prevPoolSizeAll) - (sizeDiff * buysPerVote)) / float64(prevPoolSizeAll)
	}

	// Protect pool size from going over target.
	// Amplify targetRatio by intervals over pool target.
	if targetRatio > 1.0 {
		sizeDiff := float64(curPoolSizeAll - targetPoolSizeAll)
		relativeIntervals := sizeDiff / float64(ticketsPerWindow)
		targetRatio = (float64(targetPoolSizeAll) + (sizeDiff * relativeIntervals)) / float64(targetPoolSizeAll)
	}

	// Voila!
	nextDiff := float64(curDiff) * poolSizeChangeRatio * targetRatio

	// Ramp up, optimize for more staked coins and better price discovery.
	// Detect for below target pool size with pool size increasing.
	// Amplify poolSizeChangeRatio by freshStakeRatio.
	// With the poolSizeChangeRatio increasing, freshStakeRatio will naturally be over 1.
	if curPoolSizeAll < targetPoolSizeAll-maxFreshStakePerWindow && poolSizeChangeRatio > 1.0 {
		poolSizeDiff := float64(curPoolSizeAll - prevPoolSizeAll)
		poolSizeChangeRatio = (float64(prevPoolSizeAll) + (poolSizeDiff * freshStakeRatio)) / float64(prevPoolSizeAll)
		nextDiff = float64(curDiff) * poolSizeChangeRatio // exclude targetRatio
	}

	// ramp up price during initial pool population
	maximumStakeDiff := int64(float64(s.tip.totalSupply) / float64(targetPoolSize))
	if int64(nextDiff) > maximumStakeDiff && targetRatio < 1.0 {
		nextDiff = float64(maximumStakeDiff) * targetRatio
	}

	// optional
	if int64(nextDiff) > maximumStakeDiff {
		if maximumStakeDiff < s.params.MinimumStakeDiff {
			return s.params.MinimumStakeDiff
		}
		return maximumStakeDiff
	}

	// Hard coded minimum value.
	if int64(nextDiff) < s.params.MinimumStakeDiff {
		return s.params.MinimumStakeDiff
	}

	return int64(nextDiff)
}

// the algorithm proposed by raedah (again)
func (s *simulator) calcNextStakeDiffProposal1H() int64 {
	// Stake difficulty before any tickets could possibly be purchased is
	// the minimum value.
	nextHeight := int32(0)
	if s.tip != nil {
		nextHeight = s.tip.height + 1
	}
	stakeDiffStartHeight := int32(s.params.CoinbaseMaturity) + 1
	if nextHeight < stakeDiffStartHeight {
		return s.params.MinimumStakeDiff
	}

	// Return the previous block's difficulty requirements if the next block
	// is not at a difficulty retarget interval.
	intervalSize := s.params.StakeDiffWindowSize
	curDiff := s.tip.ticketPrice
	if int64(nextHeight)%intervalSize != 0 {
		return curDiff
	}

	// Attempt to get the pool size from the previous retarget interval.
	var prevPoolSize int64
	prevRetargetHeight := nextHeight - int32(intervalSize)
	node := s.ancestorNode(s.tip, prevRetargetHeight, nil)
	if node != nil {
		prevPoolSize = int64(node.poolSize)
	}

	// get the immature ticket count from the previous window
	// note, make sure we have no off-by-ones here
	var prevImmatureTickets int64
	ticketMaturity := int64(s.params.TicketMaturity)
	relevantHeight := s.tip.height - int32(intervalSize) // or nextHeight?
	relevantNode := s.ancestorNode(s.tip, relevantHeight, nil)
	s.ancestorNode(relevantNode, relevantHeight-int32(ticketMaturity), func(n *blockNode) {
		prevImmatureTickets += int64(len(n.ticketsAdded))
	})

	// Return the existing ticket price for the first interval.
	if prevPoolSize+prevImmatureTickets == 0 {
		return curDiff
	}

	// derive ratio of percent change in pool size
	immatureTickets := int64(len(s.immatureTickets))
	curPoolSize := int64(s.tip.poolSize)
	curPoolSizeAll := curPoolSize + immatureTickets
	prevPoolSizeAll := prevPoolSize + prevImmatureTickets

	// derive ratio of percent of target pool size
	ticketsPerBlock := int64(s.params.TicketsPerBlock)
	ticketPoolSize := int64(s.params.TicketPoolSize)
	targetPoolSize := ticketsPerBlock * ticketPoolSize
	targetPoolSizeAll := ticketsPerBlock * (ticketPoolSize + ticketMaturity)
	targetRatio := float64(curPoolSizeAll) / float64(targetPoolSizeAll)

	// Best
	poolSizeChangeRatio := float64(curPoolSizeAll) / float64(prevPoolSizeAll)
	nextDiff := float64(curDiff) * poolSizeChangeRatio * targetRatio
	ticketsPerWindow := ticketsPerBlock * intervalSize
	maxFreshStakePerBlock := int64(s.params.MaxFreshStakePerBlock)
	if poolSizeChangeRatio < 1.0 {
		// Upward price movements are stronger then downward movements.
		// Add downward movements relative strength, for the market to respond and give its input.
		maxFreshStakePerWindow := maxFreshStakePerBlock * intervalSize
		buysPerVote := float64(maxFreshStakePerWindow) / float64(ticketsPerWindow)
		sizeDiff := float64(prevPoolSizeAll) - float64(curPoolSizeAll)
		tempPoolSizeChangeRatio := (float64(prevPoolSizeAll) - (sizeDiff * buysPerVote)) / float64(prevPoolSizeAll)
		nextDiff = float64(curDiff) * tempPoolSizeChangeRatio * targetRatio
	} else {
		sizeDiff := float64(curPoolSizeAll - prevPoolSizeAll)
		relativeIntervals := sizeDiff / float64(intervalSize)
		tempPoolSizeChangeRatio := (float64(prevPoolSizeAll) + (sizeDiff * relativeIntervals)) / float64(prevPoolSizeAll)
		nextDiff = float64(curDiff) * tempPoolSizeChangeRatio * targetRatio
	}

	/*
	// becomes wavier
	//sizeDiff := math.Abs(float64(curPoolSizeAll) - float64(prevPoolSizeAll))
	poolSizeChangeRatio := float64(curPoolSizeAll) / float64(prevPoolSizeAll)
	var relativeMultiplier float64
	if poolSizeChangeRatio < 1.0 {
	    sizeDiff := float64(prevPoolSizeAll) - float64(curPoolSizeAll)
	    relativeIntervals := sizeDiff / float64(intervalSize)
	    maxFreshStakePerBlock := int64(s.params.MaxFreshStakePerBlock)
	    votesPerBuy := float64(ticketsPerBlock) / float64(maxFreshStakePerBlock)
	    //relativeIntervals = relativeIntervals * votesPerBuy
	    relativeMultiplier = (float64(prevPoolSizeAll) + (sizeDiff * relativeIntervals * votesPerBuy)) / float64(prevPoolSizeAll)
	} else {
	    sizeDiff := float64(curPoolSizeAll) - float64(prevPoolSizeAll)
	    relativeIntervals := sizeDiff / float64(intervalSize)
	    relativeMultiplier = (float64(prevPoolSizeAll) + (sizeDiff * relativeIntervals)) / float64(prevPoolSizeAll)
	}
	nextDiff := float64(curDiff) * relativeMultiplier * targetRatio
	*/

	/*
        // continous waves
        poolShift := math.Abs(float64(curPoolSizeAll - prevPoolSizeAll))
        maxFreshStakePerBlock := int64(s.params.MaxFreshStakePerBlock)
        var relativeIntervals float64
        if curPoolSizeAll < prevPoolSizeAll {
                relativeIntervals = float64(maxFreshStakePerBlock) / float64(ticketsPerBlock)
        } else {
                relativeIntervals = poolShift / float64(intervalSize)
        }
        relativeMultiplier := (float64(prevPoolSizeAll) + (poolShift * relativeIntervals)) / float64(prevPoolSizeAll)
        nextDiff := float64(curDiff) * relativeMultiplier * targetRatio
	*/

	// ramp up price during initial pool population
	maximumStakeDiff := int64(float64(s.tip.totalSupply) / float64(targetPoolSize))
	if int64(nextDiff) > maximumStakeDiff && targetRatio < 1.0 {
		nextDiff = float64(maximumStakeDiff) * targetRatio
	}

	// optional
	if int64(nextDiff) > maximumStakeDiff {
		if maximumStakeDiff < s.params.MinimumStakeDiff {
			return s.params.MinimumStakeDiff
		}
		return maximumStakeDiff
	}

	// hard coded minimum value
	if int64(nextDiff) < s.params.MinimumStakeDiff {
		return s.params.MinimumStakeDiff
	}

	return int64(nextDiff)
}

// calcNextStakeDiffProposal2 returns the required stake difficulty (aka ticket
// price) for the block after the current tip block the simulator is associated
// with using the algorithm proposed by animedow in
// https://github.com/decred/dcrd/issues/584
func (s *simulator) calcNextStakeDiffProposal2() int64 {
	// Stake difficulty before any tickets could possibly be purchased is
	// the minimum value.
	nextHeight := int32(0)
	if s.tip != nil {
		nextHeight = s.tip.height + 1
	}
	stakeDiffStartHeight := int32(s.params.CoinbaseMaturity) + 1
	if nextHeight < stakeDiffStartHeight {
		return s.params.MinimumStakeDiff
	}

	// Return the previous block's difficulty requirements if the next block
	// is not at a difficulty retarget interval.
	intervalSize := s.params.StakeDiffWindowSize
	curDiff := s.tip.ticketPrice
	if int64(nextHeight)%intervalSize != 0 {
		return curDiff
	}

	//                ax
	// f(x) = - ---------------- + d
	//           (x - b)(x + c)
	//
	// x = amount of ticket deviation from the target pool size;
	// a = a modifier controlling the slope of the function;
	// b = the maximum boundary;
	// c = the minimum boundary;
	// d = the average ticket price in pool.
	x := int64(s.tip.poolSize) - (int64(s.params.TicketsPerBlock) *
		int64(s.params.TicketPoolSize))
	a := int64(100000)
	b := int64(2880)
	c := int64(2880)
	var d int64
	var totalSpent int64
	totalTickets := int64(len(s.immatureTickets) + s.liveTickets.Len())
	if totalTickets != 0 {
		for _, ticket := range s.immatureTickets {
			totalSpent += int64(ticket.price)
		}
		s.liveTickets.ForEach(func(k tickettreap.Key, v *tickettreap.Value) bool {
			totalSpent += v.PurchasePrice
			return true
		})
		d = totalSpent / totalTickets
	}
	price := int64(float64(d) - 100000000*(float64(a*x)/float64((x-b)*(x+c))))
	if price < s.params.MinimumStakeDiff {
		price = s.params.MinimumStakeDiff
	}
	return price
}

// calcNextStakeDiffProposal3 returns the required stake difficulty (aka ticket
// price) for the block after the current tip block the simulator is associated
// with using the algorithm proposed by coblee in
// https://github.com/decred/dcrd/issues/584
func (s *simulator) calcNextStakeDiffProposal3() int64 {
	// Stake difficulty before any tickets could possibly be purchased is
	// the minimum value.
	nextHeight := int32(0)
	if s.tip != nil {
		nextHeight = s.tip.height + 1
	}
	stakeDiffStartHeight := int32(s.params.CoinbaseMaturity) + 1
	if nextHeight < stakeDiffStartHeight {
		return s.params.MinimumStakeDiff
	}

	// Return the previous block's difficulty requirements if the next block
	// is not at a difficulty retarget interval.
	intervalSize := s.params.StakeDiffWindowSize
	curDiff := s.tip.ticketPrice
	if int64(nextHeight)%intervalSize != 0 {
		return curDiff
	}

	// f(x) = x*(locked/target_pool_size) + (1-x)*(locked/pool_size_actual)
	ticketsPerBlock := int64(s.params.TicketsPerBlock)
	targetPoolSize := ticketsPerBlock * int64(s.params.TicketPoolSize)
	lockedSupply := s.tip.stakedCoins
	x := int64(1)
	var price int64
	if s.tip.poolSize == 0 {
		price = int64(lockedSupply) / targetPoolSize
	} else {
		price = x*int64(lockedSupply)/targetPoolSize +
			(1-x)*(int64(lockedSupply)/int64(s.tip.poolSize))
	}
	if price < s.params.MinimumStakeDiff {
		price = s.params.MinimumStakeDiff
	}
	return price
}

// calcNextStakeDiffProposal4 returns the required stake difficulty (aka ticket
// price) for the block after the current tip block the simulator is associated
// with using the algorithm proposed by jyap808 in
// https://github.com/decred/dcrd/issues/584
func (s *simulator) calcNextStakeDiffProposal4() int64 {
	// Stake difficulty before any tickets could possibly be purchased is
	// the minimum value.
	nextHeight := int32(0)
	if s.tip != nil {
		nextHeight = s.tip.height + 1
	}

	stakeDiffStartHeight := int32(s.params.CoinbaseMaturity) + 1
	if nextHeight < stakeDiffStartHeight {
		return s.params.MinimumStakeDiff
	}

	// Return the previous block's difficulty requirements if the next block
	// is not at a difficulty retarget interval.
	intervalSize := s.params.StakeDiffWindowSize
	curDiff := s.tip.ticketPrice
	if int64(nextHeight)%intervalSize != 0 {
		return curDiff
	}

	// Get the number of tickets purchased in the previous interval.
	var ticketsPurchased int64
	prevRetargetHeight := s.tip.height - int32(intervalSize)
	s.ancestorNode(s.tip, prevRetargetHeight, func(node *blockNode) {
		ticketsPurchased += int64(len(node.ticketsAdded))
	})

	// Shorter versions of useful params for convenience.
	votesPerBlock := int64(s.params.TicketsPerBlock)
	votesPerInterval := votesPerBlock * int64(s.params.TicketPoolSize)
	maxTicketsPerBlock := int64(s.params.MaxFreshStakePerBlock)
	maxTicketsPerInterval := maxTicketsPerBlock * int64(s.params.TicketPoolSize)
	targetPoolSize := votesPerBlock * int64(s.params.TicketPoolSize)

	// Formulas provided by proposal.
	//
	// Bounds = TickPrice *  TickVotesCycle / MaxTickCycle
	// ScalingFactor = (TickBought - TickVotesCycle) / (MaxTickCycle - TickVotesCycle)
	//
	// If PoolTarget >= PoolTickets:
	//   NewTickPrice = TickPrice + (Bounds * Scaling Factor)
	// Else:
	//   NewTickPrice = TickPrice + (-Bounds * Scaling Factor)
	//
	var nextDiff int64
	bounds := float64(curDiff) * float64(votesPerInterval) /
		float64(maxTicketsPerInterval)
	scalingFactor := float64(ticketsPurchased-votesPerInterval) /
		float64(maxTicketsPerInterval-votesPerInterval)
	if targetPoolSize >= int64(s.tip.poolSize) {
		nextDiff = int64(float64(curDiff) + (bounds * scalingFactor))
	} else {
		nextDiff = int64(float64(curDiff) + (-bounds * scalingFactor))
	}

	if nextDiff < s.params.MinimumStakeDiff {
		nextDiff = s.params.MinimumStakeDiff
	}
	return nextDiff
}

var integral = 0.0
var previousError = 0.0

// calcNextStakeDiffProposal5 returns the required stake difficulty (aka ticket
// price) for the block after the current tip block the simulator is associated
// with using the algorithm proposed by edsonbrusque in
// https://github.com/decred/dcrd/issues/584
func (s *simulator) calcNextStakeDiffProposal5() int64 {
	// Stake difficulty before any tickets could possibly be purchased is
	// the minimum value.
	nextHeight := int32(0)
	if s.tip != nil {
		nextHeight = s.tip.height + 1
	}

	stakeDiffStartHeight := int32(s.params.CoinbaseMaturity) + 1
	if nextHeight < stakeDiffStartHeight {
		return s.params.MinimumStakeDiff
	}

	// Return the previous block's difficulty requirements if the next block
	// is not at a difficulty retarget interval.
	intervalSize := s.params.StakeDiffWindowSize
	curDiff := s.tip.ticketPrice
	if int64(nextHeight)%intervalSize != 0 {
		return curDiff
	}

	ticketsPerBlock := int64(s.params.TicketsPerBlock)
	targetPoolSize := ticketsPerBlock * int64(s.params.TicketPoolSize)

	Kp := 0.0017
	Ki := 0.00005
	Kd := 0.0024
	e := float64(int64(s.tip.poolSize) - targetPoolSize)
	integral = integral + e
	derivative := (e - previousError)
	nextDiff := int64(dcrutil.AtomsPerCoin * (e*Kp + integral*Ki + derivative*Kd))
	previousError = e

	if nextDiff < s.params.MinimumStakeDiff {
		nextDiff = s.params.MinimumStakeDiff
	}
	return nextDiff
}
