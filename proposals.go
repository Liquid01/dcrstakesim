// Copyright (c) 2017 Dave Collins
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/davecgh/dcrstakesim/internal/tickettreap"
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

	// Attempt to get the ticket price and pool size from the previous
	// retarget interval.
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

// the algorithm proposed by raedah (enhanced)
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

	// Attempt to get the ticket price and pool size from the previous
	// retarget interval.
	prevRetargetHeight := nextHeight - int32(intervalSize)
	prevPoolSize := int64(s.tip.poolSize)
	//prevDiff := s.tip.ticketPrice
	node := s.ancestorNode(s.tip, prevRetargetHeight, nil)
	if node != nil {
		prevPoolSize = int64(node.poolSize)
		//prevDiff = node.ticketPrice
	}

	// get the total amount of tickets purchased including
	// the immature ones
	/*
	   var totalPurchased int64
	   prevRetargetHeight := nextHeight - int32(intervalSize)
	   node := s.ancestorNode(s.tip, prevRetargetHeight, func(n *blockNode) {
	       totalPurchased += int64(len(n.ticketsAdded))
	   })
	*/

	// Return the existing ticket price for the first interval.
	if prevPoolSize == 0 {
		return curDiff
	}

	// derive ratio of percent change in pool size
	// max possible poolSizeChangeRatio is 2
	curPoolSize := int64(s.tip.poolSize)
	immatureTickets := int64(len(s.immatureTickets))
	curPoolSizeAll := curPoolSize + immatureTickets
	poolSizeChangeRatio := float64(curPoolSizeAll) / float64(prevPoolSize)

	/*
			// derive ratio of purchase slots filled
		        maxFreshStakePerBlock := int64(s.params.MaxFreshStakePerBlock)
		        //blocksPerWindow := int64(s.params.BlocksPerWindow)
		        blocksPerWindow := int64(144)
		        maxFreshStakePerWindow := maxFreshStakePerBlock * blocksPerWindow
		        freshStakeLastWindow := curPoolSize - prevPoolSize
		        freshStakeRatio := freshStakeLastWindow / maxFreshStakePerWindow
	*/

	// derive ratio of percent of target pool size
	ticketsPerBlock := int64(s.params.TicketsPerBlock)
	targetPoolSize := ticketsPerBlock * int64(s.params.TicketPoolSize)
	targetRatio := float64(curPoolSizeAll) / float64(targetPoolSize)

	// s.totalSupply

	/*
	   ticketsNeeded := targetPoolSize - curPoolSize

	   //curDiff prevDiff
	   //prevPoolSize curPoolSize targetPoolSize

	   // difference in price from previous window and current window
	   diffRatio := prevDiff / curDiff
	*/

	// create var
	nextDiff := float64(curDiff)

	/*
	   // ticket pool is not full enough yet
	   if ticketsNeeded > maxFreshStakePerWindow {
	       nextDiff = float64(curDiff) * (ticketsPerBlock / maxFreshStakePerBlock)
	       if int64(nextDiff) < s.params.MinimumStakeDiff {
	               return s.params.MinimumStakeDiff
	       }
	   }
	*/

	//if ticketsNeeded > freshStakeLastWindow {

	/*
	   // there was not a price change last window
	   if priceRatio == 1.0 {
	       // adjust the price relative to the change in pool size that occured
	       nextDiff = float64(curDiff) * poolSizeChangeRatio
	   }

	   if priceRatio > 1.0 {
	       if ticketsNeeded > maxFreshStakePerWindow {
	   }
	*/

	/*
	   if priceRatio > 1.0 {
	   // price went up last window
	       // we are under the target pool size
	       if targetRatio < 1.0 {
	       }
	       // we are over the target pool size
	       if targetRatio > 1.0 {
	       }
	       // we are at the target pool size
	       if targetRatio == 1.0 {
	           // price stays the same
	           nextDiff = float64(curDiff)
	       }
	   }

	   if priceRatio < 1.0 {
	   // price went down last window
	   }

	   /*
	   if poolSizeChangeRatio < 1.0 {  // if pool size decreased last window
	       nextDiff = float64(curDiff) * (1.0 - ((1.0 - poolSizeChangeRatio) * (1/targetRatio)))
	   }
	   if poolSizeChangeRatio > 1.0 {  // if pool size increased last window
	       nextDiff = float64(curDiff) * (1.0 + ((poolSizeChangeRatio - 1.0) * targetRatio))
	   }
	   if poolSizeChangeRatio == 1.0 {  // if pool size stayed the same last window
	       nextDiff = float64(curDiff) * targetRatio
	   }
	*/

	/*
		//pscrRate is the poolSizeChangeRatio with a multiplying factor
		factorPSCR := 1.0
		pscrRate := 1.0
		if poolSizeChangeRatio < 1.0 {
			pscrRate = 1.0 - ((1.0 - poolSizeChangeRatio) * factorPSCR)
		}
		if poolSizeChangeRatio > 1.0 {
			pscrRate = 1.0 + ((poolSizeChangeRatio - 1.0) * factorPSCR)
		}
		// what about if == 1.0 ?

		//trRate is the amount we should adjust the price relative
		//to the target pool size
		factorTR := 1.0
		trRate := 1.0
		if targetRatio < 1.0 {
			trRate = 1 / trRate
			trRate = ((1.0 - trRate) * factorTR) + 1.0
			// below pool target amount so decrease price by this much
		}
		if targetRatio > 1.0 {
			trRate = ((1.0 - targetRatio) * factorTR) + 1.0
			trRate = 1 / trRate
			// above pool target amount so increase price by this much
		}
		// what about if == 1.0
	*/

	/*
		if targetRatio < 1.0 {
			nextDiff = float64(curDiff) * (pscrRate * targetRatio)
		}
		if targetRatio > 1.0 {
			nextDiff = float64(curDiff) + 10e8
			//nextDiff = float64(curDiff) * float64(freshStakeRatio * (maxFreshStakePerBlock/ticketsPerBlock) )
			/ *
			   if freshStakeRatio > (ticketsPerBlock / maxFreshStakePerBlock) {
			       nextDiff = float64(curDiff) * (targetRatio * 10)
			   } else {
			       nextDiff = float64(curDiff) * pscrRate
			   }
			* /
		}
		// what if == 1.0
	*/

	nextDiff = float64(curDiff) * (poolSizeChangeRatio * targetRatio)

	/*
	   factor := 1.0 // increase the action
	   if poolSizeChangeRatio < 1.0 {
	       nextDiff = float64(curDiff) * (1.0 - (((1.0 - poolSizeChangeRatio) * factor) * (1/targetRatio)))
	   }
	   if poolSizeChangeRatio > 1.0 {
	       nextDiff = float64(curDiff) * (1.0 + (((poolSizeChangeRatio - 1.0) * factor) * targetRatio))
	   }
	*/

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
	lockedSupply := s.totalSupply - s.spendableSupply
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
