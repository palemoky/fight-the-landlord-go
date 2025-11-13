package rule

import (
	"github.com/palemoky/fight-the-landlord-go/internal/card"
)

// parseRocket 王炸
func parseRocket(analysis HandAnalysis, cards []card.Card) (ParsedHand, bool) {
	if len(cards) == 2 && analysis.counts[card.RankBlackJoker] == 1 && analysis.counts[card.RankRedJoker] == 1 {
		return ParsedHand{Type: Rocket, KeyRank: card.RankRedJoker, Cards: cards}, true
	}
	return ParsedHand{}, false
}

// parseBomb 炸弹
func parseBomb(analysis HandAnalysis, cards []card.Card) (ParsedHand, bool) {
	if len(analysis.fours) == 1 && len(cards) == 4 {
		return ParsedHand{Type: Bomb, KeyRank: analysis.fours[0], Cards: cards}, true
	}
	return ParsedHand{}, false
}

// parseFourWithKickers 四带二
func parseFourWithKickers(analysis HandAnalysis, cards []card.Card) (ParsedHand, bool) {
	cardLen := len(cards)
	if len(analysis.fours) == 1 {
		hand := ParsedHand{KeyRank: analysis.fours[0], Cards: cards}
		if cardLen == 6 && (len(analysis.ones) == 2 || len(analysis.pairs) == 1) { // AAAABB、AAAABC
			// 四带二，可以带两张单牌，也可以带一个对子(不算四带两对)
			hand.Type = FourWithTwo
			return hand, true
		}
		if cardLen == 8 && len(analysis.pairs) == 2 { // AAAABBCC、AAAABBBB
			hand.Type = FourWithTwoPairs
			return hand, true
		}
	}
	return ParsedHand{}, false
}

// parseTrioWithKickers 三带X
func parseTrioWithKickers(analysis HandAnalysis, cards []card.Card) (ParsedHand, bool) {
	cardLen := len(cards)
	if len(analysis.trios) == 1 {
		hand := ParsedHand{KeyRank: analysis.trios[0], Cards: cards}
		if cardLen == 4 && len(analysis.ones) == 1 { // AAAB
			hand.Type = TrioWithSingle
			return hand, true
		}
		if cardLen == 5 && len(analysis.pairs) == 1 { // AAABB
			hand.Type = TrioWithPair
			return hand, true
		}
	}
	return ParsedHand{}, false
}

// parsePlane 飞机
func parsePlane(analysis HandAnalysis, cards []card.Card) (ParsedHand, bool) {
	cardLen, planeLen := len(cards), len(analysis.trios)
	if isContinuous(analysis.trios) && planeLen >= 2 {
		hand := ParsedHand{KeyRank: analysis.trios[0], Length: planeLen, Cards: cards}
		// 飞机不带翅膀
		if planeLen*3 == cardLen { // AAABBB+
			hand.Type = Plane
			return hand, true
		}
		// 飞机带单
		if planeLen*4 == cardLen && len(analysis.ones) == planeLen { // AAABBBCD+、AAABBAC+、AAABBBCC+
			hand.Type = PlaneWithSingles
			return hand, true
		}
		// 飞机带对
		if planeLen*5 == cardLen && len(analysis.pairs) == planeLen { // AAABBBCCDD+
			hand.Type = PlaneWithPairs
			return hand, true
		}
	}
	return ParsedHand{}, false
}

// parseStraight 顺子
func parseStraight(analysis HandAnalysis, cards []card.Card) (ParsedHand, bool) {
	cardLen := len(cards)
	if isContinuous(analysis.ones) && len(analysis.ones) == cardLen && cardLen >= 5 { // ABCDE+
		return ParsedHand{Type: Straight, KeyRank: analysis.ones[0], Length: cardLen, Cards: cards}, true
	}
	return ParsedHand{}, false
}

// parsePairStraight 连对
func parsePairStraight(analysis HandAnalysis, cards []card.Card) (ParsedHand, bool) {
	if isContinuous(analysis.pairs) && len(analysis.pairs)*2 == len(cards) && len(analysis.pairs) >= 3 { // AABBCC+
		return ParsedHand{Type: PairStraight, KeyRank: analysis.pairs[0], Length: len(analysis.pairs), Cards: cards}, true
	}
	return ParsedHand{}, false
}

// parseSimpleType 简单牌型：单、对、三
func parseSimpleType(analysis HandAnalysis, cards []card.Card) (ParsedHand, bool) {
	if len(analysis.counts) == 1 {
		hand := ParsedHand{KeyRank: analysis.ones[0], Cards: cards}
		switch len(cards) {
		case 1:
			hand.Type = Single
			return hand, true
		case 2:
			hand.Type = Pair
			return hand, true
		case 3:
			hand.Type = Trio
			return hand, true
		}
	}
	return ParsedHand{}, false
}
