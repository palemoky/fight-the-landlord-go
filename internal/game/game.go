package game

import (
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
	Players           [3]*Player      // 玩家
	Deck              card.Deck       // 手牌
	LandlordCards     []card.Card     // 地主手牌
	CurrentTurn       int             // 当前出牌玩家
	LastPlayedHand    rule.ParsedHand // 上家出牌
	LastPlayerIdx     int             // 上家
	ConsecutivePasses int
	CardCounter       *card.CardCounter
	ui                UI
}

// UI 是一个接口，定义了游戏与用户交互的所有方法
type UI interface {
	DisplayGame(*Game)
	GetPlayerInput(*Player, time.Duration) (string, bool)
	ShowMessage(string)
	ShowError(error)
	ClearScreen()
	DisplayRules()
}

func NewGame(ui UI) *Game {
	players := [3]*Player{
		{Name: "Player 1"},
		{Name: "Player 2"},
		{Name: "Player 3"},
	}
	deck := card.NewDeck()
	deck.Shuffle()

	return &Game{
		Players:     players,
		Deck:        deck,
		CardCounter: card.NewCardCounter(),
		ui:          ui,
	}
}

func (g *Game) Deal() {
	for range 17 {
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

func (g *Game) Bidding() {
	g.ui.ClearScreen()
	g.ui.ShowMessage("--- 叫地主阶段 ---")
	bidderIdx := rand.Intn(3)
	fmt.Printf("从 %s 开始叫地主。\n", g.Players[bidderIdx].Name)

	time.Sleep(2 * time.Second)

	g.Players[bidderIdx].IsLandlord = true
	g.Players[bidderIdx].Hand = append(g.Players[bidderIdx].Hand, g.LandlordCards...)
	g.Players[bidderIdx].SortHand()

	g.CurrentTurn = bidderIdx
	g.LastPlayerIdx = bidderIdx

	g.ui.ShowMessage(fmt.Sprintf("%s 成为地主！并获得三张底牌。\n", g.Players[bidderIdx].Name))
	g.ui.ShowMessage(fmt.Sprintf("底牌是: %s", g.LandlordCards))
	// time.Sleep(3 * time.Second)
}

func (g *Game) IsLandlordCardPlayed(c card.Card) bool {
	// 获取当前场上该点数牌的剩余数量
	remainingCount := g.CardCounter.GetRemainingCards()[c.Rank]

	// 定义一副牌中每个点数的初始数量
	initialCount := 4
	if c.Rank == card.RankBlackJoker || c.Rank == card.RankRedJoker {
		initialCount = 1
	}

	// 如果剩余数量 < 初始数量，意味着至少有一张该点数的牌被打出
	// 对于UI显示来说，这足以作为“灰度化”底牌的依据
	return remainingCount < initialCount
}

func (g *Game) Run() {
	g.Deal()
	g.Bidding()

	for {
		g.ui.DisplayGame(g)
		currentPlayer := g.Players[g.CurrentTurn]

		// 1. 调用带有倒计时功能的 GetPlayerInput
		input, timedOut := g.ui.GetPlayerInput(currentPlayer, PlayerTurnTimeout)

		upperInput := strings.ToUpper(input)
		if upperInput == "HELP" || upperInput == "RULES" {
			g.ui.DisplayRules() // 调用UI来显示规则
			continue            // 跳过本轮的后续逻辑，重新渲染游戏界面
		}

		// 5. 根据是否超时来处理输入
		if timedOut {
			// 如果是轮到你自由出牌，不能pass，自动打出最小的单牌
			if g.LastPlayerIdx == g.CurrentTurn || g.LastPlayedHand.IsEmpty() {
				g.ui.ShowMessage("系统自动为您打出最小的一张牌。")
				// 查找并“模拟输入”最小的牌
				minCard := currentPlayer.Hand[len(currentPlayer.Hand)-1] // 因为手牌是排序的，最后一张就是最小的
				input = minCard.Rank.String()
			} else {
				g.ui.ShowMessage("系统自动为您选择“要不起”。")
				input = "PASS"
			}
		}

		if input == "PASS" {
			if g.LastPlayerIdx == g.CurrentTurn {
				fmt.Println("你是第一个出牌的，不能 PASS！")
				time.Sleep(2 * time.Second)
				continue
			}
			g.ConsecutivePasses++
			if g.ConsecutivePasses == 2 {
				fmt.Println("所有人都 PASS，轮到上一轮出牌者重新出牌。")
				g.LastPlayedHand = rule.ParsedHand{} // 使用 rule.ParsedHand
				g.ConsecutivePasses = 0
				g.LastPlayerIdx = g.CurrentTurn // 轮到上个出牌者继续出牌
			}
			g.CurrentTurn = (g.CurrentTurn + 1) % 3
			continue
		}

		cardsToPlay, err := rule.FindCardsInHand(currentPlayer.Hand, input)
		if err != nil {
			fmt.Println("出牌无效: ", err)
			time.Sleep(2 * time.Second)
			continue
		}

		handToPlay, err := rule.ParseHand(cardsToPlay)
		if err != nil {
			fmt.Println("无效的牌型: ", err)
			time.Sleep(2 * time.Second)
			continue
		}

		canPlay := false
		if g.LastPlayerIdx == g.CurrentTurn || g.LastPlayedHand.IsEmpty() {
			canPlay = true
		} else {
			canPlay = rule.CanBeat(handToPlay, g.LastPlayedHand)
		}

		if canPlay {
			g.LastPlayedHand = handToPlay
			g.LastPlayerIdx = g.CurrentTurn
			g.ConsecutivePasses = 0

			g.CardCounter.Update(cardsToPlay) // 出牌后更新记牌器
			currentPlayer.Hand = rule.RemoveCards(currentPlayer.Hand, cardsToPlay)

			if len(currentPlayer.Hand) == 0 {

				g.ui.DisplayGame(g)
				fmt.Println("\n================== 游戏结束 ==================")
				if currentPlayer.IsLandlord {
					fmt.Printf("🎉 地主 (%s) 获胜!\n", currentPlayer.Name)
				} else {
					fmt.Printf("🥳 农民 (%s) 获胜!\n", currentPlayer.Name)
				}
				return
			}
			g.CurrentTurn = (g.CurrentTurn + 1) % 3
		} else {
			fmt.Println("你的牌没有大过上家!")
			time.Sleep(2 * time.Second)
		}
	}
}
