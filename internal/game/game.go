package game

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/palemoky/fight-the-landlord-go/internal/card"
	"github.com/palemoky/fight-the-landlord-go/internal/rule"
)

const (
	PlayerTurnTimeout = 30 * time.Second
)

// Game 定义游戏状态
type Game struct {
	Players           [3]*Player
	Deck              card.Deck
	LandlordCards     []card.Card
	CurrentTurn       int
	LastPlayedHand    rule.ParsedHand
	LastPlayerIdx     int
	ConsecutivePasses int
	CardCounter       *card.CardCounter
}

// NewGame 初始化一个新游戏
func NewGame() *Game {
	players := [3]*Player{
		{Name: "Player 1 (你)"},
		{Name: "Player 2"},
		{Name: "Player 3"},
	}
	deck := card.NewDeck()
	deck.Shuffle()

	return &Game{
		Players:     players,
		Deck:        deck,
		CardCounter: card.NewCardCounter(),
	}
}

// Deal 发牌
func (g *Game) Deal() {
	for i := 0; i < 17; i++ {
		for _, p := range g.Players {
			p.Hand = append(p.Hand, g.Deck[0])
			g.Deck = g.Deck[1:]
		}
	}
	g.LandlordCards = g.Deck
	for _, p := range g.Players {
		p.SortHand()
	}
}

// Bidding 叫地主（此处为简化版，随机选择一个）
func (g *Game) Bidding() {
	landlordIdx := rand.Intn(3)
	g.Players[landlordIdx].IsLandlord = true
	g.Players[landlordIdx].Hand = append(g.Players[landlordIdx].Hand, g.LandlordCards...)
	g.Players[landlordIdx].SortHand()

	g.CurrentTurn = landlordIdx
	g.LastPlayerIdx = landlordIdx
}

// PlayTurn 处理玩家的一次出牌操作
func (g *Game) PlayTurn(input string) error {
	currentPlayer := g.Players[g.CurrentTurn]
	isTimeout := input == ""

	// 处理超时：如果轮到你自由出牌，则自动出最小的牌；否则自动PASS
	if isTimeout {
		if g.LastPlayerIdx == g.CurrentTurn || g.LastPlayedHand.IsEmpty() {
			minCard := currentPlayer.Hand[len(currentPlayer.Hand)-1] // 手牌已排序，最后一张最小
			input = minCard.Rank.String()
		} else {
			input = "PASS"
		}
	}

	upperInput := strings.ToUpper(strings.TrimSpace(input))

	if upperInput == "PASS" {
		if g.LastPlayerIdx == g.CurrentTurn || g.ConsecutivePasses == 2 {
			return errors.New("轮到你出牌，不能PASS")
		}
		g.ConsecutivePasses++
		if g.ConsecutivePasses == 2 {
			// 如果连续两人PASS，则开启新的一轮
			g.LastPlayedHand = rule.ParsedHand{}
			g.LastPlayerIdx = (g.CurrentTurn + 1) % 3 // 新一轮由下家开始
		}
		g.CurrentTurn = (g.CurrentTurn + 1) % 3
		return nil
	}

	cardsToPlay, err := rule.FindCardsInHand(currentPlayer.Hand, upperInput)
	if err != nil {
		return fmt.Errorf("出牌无效: %w", err)
	}

	handToPlay, err := rule.ParseHand(cardsToPlay)
	if err != nil {
		return fmt.Errorf("无效的牌型: %w", err)
	}

	isNewRound := g.LastPlayerIdx == g.CurrentTurn || g.LastPlayedHand.IsEmpty() || g.ConsecutivePasses == 2
	if isNewRound || rule.CanBeat(handToPlay, g.LastPlayedHand) {
		g.LastPlayedHand = handToPlay
		g.LastPlayerIdx = g.CurrentTurn
		g.ConsecutivePasses = 0

		g.CardCounter.Update(cardsToPlay)
		currentPlayer.Hand = rule.RemoveCards(currentPlayer.Hand, cardsToPlay)

		if len(currentPlayer.Hand) == 0 {
			// 游戏结束，由UI处理视图
			return nil
		}

		g.CurrentTurn = (g.CurrentTurn + 1) % 3
	} else {
		return errors.New("你的牌没有大过上家")
	}

	return nil
}

// CheckWinner 检查是否有玩家获胜
func (g *Game) CheckWinner() (*Player, bool) {
	for _, p := range g.Players {
		if len(p.Hand) == 0 {
			return p, true
		}
	}
	return nil, false
}
