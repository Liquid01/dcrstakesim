// Copyright (c) 2017 Dave Collins
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/davecgh/dcrstakesim/internal/tickettreap"
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
		// Increase downward price action so that it matches upward speed.
		// Upward price movements are stronger then downward movements
		// So give downward movements more relative strength
		// for the market to respond and give its input
		maxFreshStakePerBlock := int64(s.params.MaxFreshStakePerBlock)
		maxFreshStakePerWindow := maxFreshStakePerBlock * intervalSize
		thisRatio := float64(maxFreshStakePerWindow) / float64(ticketsPerWindow)
		sizeDiff := float64(prevPoolSizeAll) - float64(curPoolSizeAll)
		poolSizeChangeRatio = (float64(prevPoolSizeAll) - (sizeDiff * thisRatio)) / float64(prevPoolSizeAll)
		nextDiff = float64(curDiff) * poolSizeChangeRatio * targetRatio
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
	/*
		if int64(nextDiff) > maximumStakeDiff {
			if maximumStakeDiff < s.params.MinimumStakeDiff {
				return s.params.MinimumStakeDiff
			}
			return maximumStakeDiff
		}
	*/

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
