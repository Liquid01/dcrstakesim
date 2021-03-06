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

// The algorithm proposed by raedah (v4)
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

	// Get the immature ticket count from the previous interval.
	var prevImmatureTickets int64
	ticketMaturity := int64(s.params.TicketMaturity)
	s.ancestorNode(node, node.height-int32(ticketMaturity), func(n *blockNode) {
		prevImmatureTickets += int64(len(n.ticketsAdded))
	})

	// Return the existing ticket price for the first interval.
	if prevPoolSize+prevImmatureTickets == 0 {
		return curDiff
	}

	// Pool size amounts.
	immatureTickets := int64(len(s.immatureTickets))
	curPoolSize := int64(s.tip.poolSize)
	curPoolSizeAll := curPoolSize + immatureTickets
	prevPoolSizeAll := prevPoolSize + prevImmatureTickets

	// Ratio of the current pool size to the desired target pool size.
	ticketsPerBlock := int64(s.params.TicketsPerBlock)
	ticketPoolSize := int64(s.params.TicketPoolSize)
	targetPoolSize := ticketsPerBlock * ticketPoolSize
	targetPoolSizeAll := ticketsPerBlock * (ticketPoolSize + ticketMaturity)
	targetRatio := float64(curPoolSizeAll) / float64(targetPoolSizeAll)

	// The change in the pool size during the last window.
	poolSizeChange := math.Abs(float64(curPoolSizeAll - prevPoolSizeAll))
	// The average pool size change per block.
	poolSizeChangePerBlock := poolSizeChange / float64(intervalSize)

	// Boost price movements using the pool size change.
	var relativeBoost float64
	// poolSizeChangePerBlock is used as the multiplier, 0-20 on mainnet
	boostFactor := poolSizeChangePerBlock
	if curPoolSizeAll < prevPoolSizeAll {
		// trending down
		relativeBoost = (float64(prevPoolSizeAll) - (poolSizeChange * boostFactor)) / float64(prevPoolSizeAll)
	} else {
		// trending up or steady
		relativeBoost = (float64(prevPoolSizeAll) + (poolSizeChange * boostFactor)) / float64(prevPoolSizeAll)
	}

	// Voila!
	nextDiff := float64(curDiff) * relativeBoost * targetRatio

	// Ramp up price during initial pool population.
	maximumStakeDiff := int64(float64(s.tip.totalSupply) / float64(targetPoolSize))
	if int64(nextDiff) > maximumStakeDiff && targetRatio < 1.0 {
		nextDiff = float64(maximumStakeDiff) * targetRatio
	}

	// Optional, keep the upper bound within a set max price.
	// Trades off pool size spike, with insuring the pool gets fully populated.
	// Also keeps the chart scale more readable.
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

// The algorithm proposed by raedah (v5)
func (s *simulator) calcNextStakeDiffProposal1R() int64 {
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

	// Get the immature ticket count from the previous interval.
	var prevImmatureTickets int64
	ticketMaturity := int64(s.params.TicketMaturity)
	s.ancestorNode(node, node.height-int32(ticketMaturity), func(n *blockNode) {
		prevImmatureTickets += int64(len(n.ticketsAdded))
	})

	// Return the existing ticket price for the first interval.
	if prevPoolSize+prevImmatureTickets == 0 {
		return curDiff
	}

	// Pool size amounts.
	immatureTickets := int64(len(s.immatureTickets))
	curPoolSize := int64(s.tip.poolSize)
	curPoolSizeAll := curPoolSize + immatureTickets
	prevPoolSizeAll := prevPoolSize + prevImmatureTickets

	// Ratio of the current pool size to the desired target pool size.
	ticketsPerBlock := int64(s.params.TicketsPerBlock)
	ticketPoolSize := int64(s.params.TicketPoolSize)
	//targetPoolSize := ticketsPerBlock * ticketPoolSize
	targetPoolSizeAll := ticketsPerBlock * (ticketPoolSize + ticketMaturity)
	targetRatio := float64(curPoolSizeAll) / float64(targetPoolSizeAll)

	// The change in the pool size during the last window.
	poolSizeChange := math.Abs(float64(curPoolSizeAll - prevPoolSizeAll))
	poolSizeChangePerBlock := poolSizeChange / float64(intervalSize)
	poolSizeChangeRatio := float64(curPoolSizeAll) / float64(prevPoolSizeAll)

	// relativeBoost: Relative movement accelerator.
	// targetBalancer: Find optimal target balance.
	relativeBoost := poolSizeChangeRatio  // Default
	targetBalancer := targetRatio         // Default
	boostFactor := poolSizeChangePerBlock // Goes up to MaxFreshStakePerBlock
	targetDistance := math.Abs(float64(curPoolSizeAll - targetPoolSizeAll))
	intervalsTillImpact := targetDistance / poolSizeChange
	if curPoolSizeAll < prevPoolSizeAll {
		// trending down
		relativeBoost = (float64(curPoolSizeAll) - (poolSizeChange * boostFactor)) / float64(prevPoolSizeAll)
		targetBalancer = (float64(curPoolSizeAll) - (poolSizeChange * intervalsTillImpact)) / float64(targetPoolSizeAll)
	} else if curPoolSizeAll > prevPoolSizeAll {
		// trending up
		relativeBoost = (float64(curPoolSizeAll) + (poolSizeChange * boostFactor)) / float64(prevPoolSizeAll)
		targetBalancer = (float64(curPoolSizeAll) + (poolSizeChange * intervalsTillImpact)) / float64(targetPoolSizeAll)
	}

	// Voila!
	nextDiff := float64(curDiff) * relativeBoost * targetBalancer

	// Ramp up price during initial pool population.
	// Insures the pool gets fully populated.
	maximumStakeDiff := int64(float64(s.tip.totalSupply) / float64(targetPoolSizeAll))
	if float64(nextDiff) > float64(maximumStakeDiff)*targetRatio {
		nextDiff = float64(maximumStakeDiff) * targetRatio
	}

	// Optional, keep the upper bound within a set max price.
	// Insures the pool gets fully populated, but not needed if above code is used.
	// Keep the chart scale more readable, but can allow the pool size to rise slightly higher.
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

// calcNextStakeDiffProposal6 returns the required stake difficulty (aka ticket
// price) for the block after the current tip block the simulator is associated
// with using the algorithm proposed by chappjc in
// https://github.com/decred/dcrd/issues/584
func (s *simulator) calcNextStakeDiffProposal6() int64 {
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

	// NOTE: Code below by chappjc with a few minor modifications by davecgh
	// to avoid needing to modify the existing simulator infrastructure for
	// getting existing immature tickets.  There are some off by ones which
	// have been left in and noted since they're in the proposal code.

	// The following variables are used in the algorithm
	// Pool: p, c, t (previous, current, target)
	// Price : q, curDiff, n (previous, current, next)
	//
	// Attempt to get the pool size from the previous retarget interval.
	//
	// NOTE: This is off by one.  It should be nextHeight instead of
	// s.tip.height, but it has been left incorrect to match the original
	// proposal code.
	var p, q int64
	prevRetargetHeight := s.tip.height - int32(intervalSize)
	if node := s.ancestorNode(s.tip, prevRetargetHeight, nil); node != nil {
		p = int64(node.poolSize)
		q = node.ticketPrice
	}

	// Return the existing ticket price for the first interval.
	if p == 0 {
		return curDiff
	}

	c := int64(s.tip.poolSize) + int64(len(s.immatureTickets))
	t := int64(s.params.TicketsPerBlock) * int64(s.params.TicketPoolSize)

	// Useful ticket counts are A (-5 * 144) and B (15 * 144)
	//A := -int64(s.params.TicketsPerBlock) * intervalSize
	B := (int64(s.params.MaxFreshStakePerBlock) - int64(s.params.TicketsPerBlock)) * intervalSize
	t += 1280 // not B (1440)?

	// Pool velocity
	//
	// Get immature count from previous window and apply it to the previous
	// live count.
	//
	// NOTE: This is off by one.  It should be nextHeight instead of
	// s.tip.height, but it has been left incorrect to match the original
	// proposal code.
	var immprev int64
	ticketMaturity := int64(s.params.TicketMaturity)
	node := s.ancestorNode(s.tip, s.tip.height-int32(intervalSize), nil)
	s.ancestorNode(node, node.height-int32(ticketMaturity), func(n *blockNode) {
		immprev += int64(len(n.ticketsAdded))
	})
	p += immprev

	// Pool size change over last intervalSize blocks
	//
	// Option 1: fraction of previous:
	//poolDelta := float64(c) / float64(p)
	//
	// Option 2: fraction of max possible change
	// Pool size change is on [A,B] (i.e. [-720,1440] for mainnet)
	poolDelta := float64(c - p)
	// Compute fraction
	poolDelta = 1 + poolDelta/float64(B)/4.0
	// allow convergence
	if math.Abs(poolDelta-1) < 0.05 {
		poolDelta = 1
	}
	// no change -> 1, fall by 720 -> 0.75, increase by 1440 -> 1.25

	// Pool force (multiple of target pool size, signed)
	del := float64(c-t) / float64(t)

	// Price velocity damper (always positive) - A large change in price from
	// the previous window to the current will have the effect of attenuating
	// the price change we are computing here. This is a very simple way of
	// slowing down the price, at the expense of making the price a little jumpy
	// and slower to adapt to big events.
	//
	// Magnitude of price change as a percent of previous price.
	absPriceDeltaLast := math.Abs(float64(curDiff-q) / float64(q))
	// Mapped onto (0,1] by an exponential decay
	m := math.Exp(-absPriceDeltaLast * 2) // m = 80% ~= exp((10% delta) *-2)
	// NOTE: make this stochastic by replacing the number 8 something like
	// (rand.NewSource(s.tip.ticketPrice).Int63() >> 59)

	// Scale directional (signed) pool force with the exponentially-mapped
	// price derivative. Interpret the scalar input parameter as a percent
	// of this computed price delta.
	s1 := float64(100)
	pctChange := s1 / 100 * m * del
	n := float64(curDiff) * (1.0 + pctChange) * poolDelta

	// Enforce minimum and maximum prices.
	pMax := int64(s.tip.totalSupply) / int64(s.params.TicketPoolSize)
	price := int64(n)
	if price < s.params.MinimumStakeDiff {
		price = s.params.MinimumStakeDiff
	} else if price > pMax {
		price = pMax
	}

	return price
}

// estimateSupply returns an estimate of the coin supply for the provided block
// height.  This is primarily used in the stake difficulty algorithm and relies
// on an estimate to simplify the necessary calculations.  The actual total
// coin supply as of a given block height depends on many factors such as the
// number of votes included in every prior block (not including all votes
// reduces the subsidy) and whether or not any of the prior blocks have been
// invalidated by stakeholders thereby removing the PoW subsidy for the them.
func (s *simulator) estimateSupply(height int32) dcrutil.Amount {
	if height <= 0 {
		return 0
	}

	// Estimate the supply by calculating the full block subsidy for each
	// reduction interval and multiplying it the number of blocks in the
	// interval then adding the subsidy produced by number of blocks in the
	// current interval.
	supply := s.params.BlockOneSubsidy()
	reductions := int64(height) / s.params.SubsidyReductionInterval
	subsidy := s.params.BaseSubsidy
	for i := int64(0); i < reductions; i++ {
		supply += s.params.SubsidyReductionInterval * subsidy

		subsidy *= s.params.MulSubsidy
		subsidy /= s.params.DivSubsidy
	}
	supply += (int64(height) % s.params.SubsidyReductionInterval) * subsidy

	// Blocks 0 and 1 have special subsidy amounts that have already been
	// added above, so remove what their subsidies would have normally been
	// which were also added above.
	supply -= s.params.BaseSubsidy * 2

	return dcrutil.Amount(supply)
}

// calcNextStakeDiffProposal7 returns the required stake difficulty (aka ticket
// price) for the block after the current tip block the simulator is associated
// with using the algorithm proposed by raedah, jy-p, and davecgh in
// https://github.com/decred/dcrd/issues/584
func (s *simulator) calcNextStakeDiffProposal7() int64 {
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

	// Return the existing ticket price for the first few intervals.
	if prevPoolSize == 0 {
		return curDiff
	}

	// Get the immature ticket count from the previous interval.
	var prevImmatureTickets int64
	ticketMaturity := int64(s.params.TicketMaturity)
	s.ancestorNode(node, node.height-int32(ticketMaturity), func(n *blockNode) {
		prevImmatureTickets += int64(len(n.ticketsAdded))
	})

	// Derive ratio of percent change in pool size.
	immatureTickets := int64(len(s.immatureTickets))
	curPoolSize := int64(s.tip.poolSize)
	curPoolSizeAll := curPoolSize + immatureTickets
	prevPoolSizeAll := prevPoolSize + prevImmatureTickets
	poolSizeChangeRatio := float64(curPoolSizeAll) / float64(prevPoolSizeAll)

	// Derive ratio of percent of target pool size.
	ticketsPerBlock := int64(s.params.TicketsPerBlock)
	ticketPoolSize := int64(s.params.TicketPoolSize)
	targetPoolSizeAll := ticketsPerBlock * (ticketPoolSize + ticketMaturity)
	targetRatio := float64(curPoolSizeAll) / float64(targetPoolSizeAll)

	// Voila!
	nextDiff := int64(float64(curDiff) * poolSizeChangeRatio * targetRatio)

	// Limit the new stake difficulty between the minimum allowed stake
	// difficulty and a maximum value that is relative to the total supply.
	estimatedSupply := s.estimateSupply(nextHeight)
	maximumStakeDiff := int64(float64(estimatedSupply) / float64(ticketPoolSize))
	if nextDiff > maximumStakeDiff {
		nextDiff = maximumStakeDiff
	}
	if nextDiff < s.params.MinimumStakeDiff {
		nextDiff = s.params.MinimumStakeDiff
	}
	return nextDiff
}
