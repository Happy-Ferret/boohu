package main

import "container/heap"

type event interface {
	Rank() int
	Action(*game)
	Renew(*game, int)
}

type iEvent struct {
	Event event
	Index int
}

type eventQueue []iEvent

func (evq eventQueue) Len() int {
	return len(evq)
}

func (evq eventQueue) Less(i, j int) bool {
	return evq[i].Event.Rank() < evq[j].Event.Rank() ||
		evq[i].Event.Rank() == evq[j].Event.Rank() && evq[i].Index < evq[j].Index
}

func (evq eventQueue) Swap(i, j int) {
	evq[i], evq[j] = evq[j], evq[i]
}

func (evq *eventQueue) Push(x interface{}) {
	no := x.(iEvent)
	*evq = append(*evq, no)
}

func (evq *eventQueue) Pop() interface{} {
	old := *evq
	n := len(old)
	no := old[n-1]
	*evq = old[0 : n-1]
	return no
}

type simpleAction int

const (
	PlayerTurn simpleAction = iota
	Teleportation
	BerserkEnd
	SlowEnd
	ExhaustionEnd
	HasteEnd
	EvasionEnd
	LignificationEnd
	ConfusionEnd
	NauseaEnd
	DisabledShieldEnd
	CorrosionEnd
	DigEnd
	SwapEnd
	ShadowsEnd
	SlayEnd
	AccurateEnd
	BlockEnd
)

func (g *game) PushEvent(ev event) {
	iev := iEvent{Event: ev, Index: g.EventIndex}
	g.EventIndex++
	heap.Push(g.Events, iev)
}

func (g *game) PushAgainEvent(ev event) {
	iev := iEvent{Event: ev, Index: 0}
	heap.Push(g.Events, iev)
}

func (g *game) PopIEvent() iEvent {
	iev := heap.Pop(g.Events).(iEvent)
	return iev
}

type simpleEvent struct {
	ERank   int
	EAction simpleAction
}

func (sev *simpleEvent) Rank() int {
	return sev.ERank
}

func (sev *simpleEvent) Renew(g *game, delay int) {
	sev.ERank += delay
	if delay == 0 {
		g.PushAgainEvent(sev)
	} else {
		g.PushEvent(sev)
	}
}

func (sev *simpleEvent) Action(g *game) {
	switch sev.EAction {
	case PlayerTurn:
		g.ComputeNoise()
		g.LogNextTick = g.LogIndex
		g.AutoNext = g.AutoPlayer(sev)
		if g.AutoNext {
			g.TurnStats()
			return
		}
		g.Quit = g.ui.HandlePlayerTurn(sev)
		if g.Quit {
			return
		}
		g.TurnStats()
	case Teleportation:
		if !g.Player.HasStatus(StatusLignification) {
			g.Teleportation(sev)
		} else {
			g.Print("Lignification has prevented teleportation.")
		}
		g.Player.Statuses[StatusTele] = 0
	case BerserkEnd:
		g.Player.Statuses[StatusBerserk] = 0
		g.Player.Statuses[StatusSlow]++
		g.Player.Statuses[StatusExhausted] = 1
		g.Player.HP -= int(10 * g.Player.HP / Max(g.Player.HPMax(), g.Player.HP))
		g.PrintStyled("You are no longer berserk.", logStatusEnd)
		g.PushEvent(&simpleEvent{ERank: sev.Rank() + 90 + RandInt(30), EAction: SlowEnd})
		g.PushEvent(&simpleEvent{ERank: sev.Rank() + 270 + RandInt(60), EAction: ExhaustionEnd})
		g.ui.StatusEndAnimation()
	case SlowEnd:
		g.Player.Statuses[StatusSlow]--
		if g.Player.Statuses[StatusSlow] <= 0 {
			g.PrintStyled("You no longer feel slow.", logStatusEnd)
			g.ui.StatusEndAnimation()
		}
	case ExhaustionEnd:
		g.PrintStyled("You no longer feel exhausted.", logStatusEnd)
		g.Player.Statuses[StatusExhausted] = 0
		g.ui.StatusEndAnimation()
	case HasteEnd:
		g.Player.Statuses[StatusSwift]--
		if g.Player.Statuses[StatusSwift] == 0 {
			g.PrintStyled("You no longer feel speedy.", logStatusEnd)
			g.ui.StatusEndAnimation()
		}
	case EvasionEnd:
		g.Player.Statuses[StatusAgile]--
		if g.Player.Statuses[StatusAgile] == 0 {
			g.PrintStyled("You no longer feel agile.", logStatusEnd)
			g.ui.StatusEndAnimation()
		}
	case LignificationEnd:
		g.Player.Statuses[StatusLignification]--
		g.Player.HP -= int(10 * g.Player.HP / Max(g.Player.HPMax(), g.Player.HP))
		if g.Player.Statuses[StatusLignification] == 0 {
			g.PrintStyled("You no longer feel attached to the ground.", logStatusEnd)
			g.ui.StatusEndAnimation()
		}
	case ConfusionEnd:
		g.PrintStyled("You no longer feel confused.", logStatusEnd)
		g.Player.Statuses[StatusConfusion] = 0
		g.ui.StatusEndAnimation()
	case NauseaEnd:
		g.PrintStyled("You no longer feel sick.", logStatusEnd)
		g.Player.Statuses[StatusNausea] = 0
		g.ui.StatusEndAnimation()
	case DisabledShieldEnd:
		g.PrintStyled("You manage to dislodge the projectile from your shield.", logStatusEnd)
		g.Player.Statuses[StatusDisabledShield] = 0
		g.ui.StatusEndAnimation()
	case CorrosionEnd:
		g.Player.Statuses[StatusCorrosion]--
		if g.Player.Statuses[StatusCorrosion] == 0 {
			g.PrintStyled("Your equipment is now free from acid.", logStatusEnd)
			g.ui.StatusEndAnimation()
		}
	case DigEnd:
		g.Player.Statuses[StatusDig]--
		if g.Player.Statuses[StatusDig] == 0 {
			g.PrintStyled("You no longer feel like an earth dragon.", logStatusEnd)
			g.ui.StatusEndAnimation()
		}
	case SwapEnd:
		g.Player.Statuses[StatusSwap]--
		if g.Player.Statuses[StatusSwap] == 0 {
			g.PrintStyled("You no longer feel light-footed.", logStatusEnd)
			g.ui.StatusEndAnimation()
		}
	case ShadowsEnd:
		g.Player.Statuses[StatusShadows]--
		if g.Player.Statuses[StatusShadows] == 0 {
			g.PrintStyled("The shadows leave you.", logStatusEnd)
			g.ui.StatusEndAnimation()
			g.ComputeLOS()
			g.MakeMonstersAware()
		}
	case SlayEnd:
		if g.Player.Statuses[StatusSlay] <= 0 {
			break
		}
		g.Player.Statuses[StatusSlay]--
		if g.Player.Statuses[StatusSlay] == 0 {
			g.PrintStyled("You no longer feel extra slaying power.", logStatusEnd)
			g.ui.StatusEndAnimation()
			g.ComputeLOS()
			g.MakeMonstersAware()
		}
	case AccurateEnd:
		g.Player.Statuses[StatusAccurate]--
		if g.Player.Statuses[StatusAccurate] == 0 {
			g.PrintStyled("You no longer feel accurate.", logStatusEnd)
			g.ui.StatusEndAnimation()
		}
	case BlockEnd:
		g.Player.Blocked = false
	}
}

type monsterAction int

const (
	MonsterTurn monsterAction = iota
	MonsConfusionEnd
	MonsExhaustionEnd
	MonsSlowEnd
	MonsLignificationEnd
)

type monsterEvent struct {
	ERank   int
	NMons   int
	EAction monsterAction
}

func (mev *monsterEvent) Rank() int {
	return mev.ERank
}

func (mev *monsterEvent) Action(g *game) {
	switch mev.EAction {
	case MonsterTurn:
		mons := g.Monsters[mev.NMons]
		if mons.Exists() {
			mons.HandleTurn(g, mev)
		}
	case MonsConfusionEnd:
		mons := g.Monsters[mev.NMons]
		if mons.Exists() {
			mons.Statuses[MonsConfused] = 0
			if g.Player.LOS[mons.Pos] {
				g.Printf("The %s is no longer confused.", mons.Kind)
			}
			mons.Path = mons.APath(g, mons.Pos, mons.Target)
		}
	case MonsLignificationEnd:
		mons := g.Monsters[mev.NMons]
		if mons.Exists() {
			mons.Statuses[MonsLignified] = 0
			if g.Player.LOS[mons.Pos] {
				g.Printf("%s is no longer lignified.", mons.Kind.Definite(true))
			}
			mons.Path = mons.APath(g, mons.Pos, mons.Target)
		}
	case MonsSlowEnd:
		mons := g.Monsters[mev.NMons]
		if mons.Exists() {
			mons.Statuses[MonsSlow]--
			if g.Player.LOS[mons.Pos] {
				g.Printf("%s is no longer slowed.", mons.Kind.Definite(true))
			}
		}
	case MonsExhaustionEnd:
		mons := g.Monsters[mev.NMons]
		if mons.Exists() {
			mons.Statuses[MonsExhausted]--
			//if mons.State != Resting && g.Player.LOS[mons.Pos] &&
			//(mons.Kind.Ranged() || mons.Kind.Smiting()) && mons.Pos.Distance(g.Player.Pos) > 1 {
			//g.Printf("%s is ready to fire again.", mons.Kind.Definite(true))
			//}
		}
	}
}

func (mev *monsterEvent) Renew(g *game, delay int) {
	mev.ERank += delay
	g.PushEvent(mev)
}

type cloudAction int

const (
	CloudEnd cloudAction = iota
	ObstructionEnd
	ObstructionProgression
	FireProgression
	NightProgression
)

type cloudEvent struct {
	ERank   int
	Pos     position
	EAction cloudAction
}

func (cev *cloudEvent) Rank() int {
	return cev.ERank
}

func (cev *cloudEvent) Action(g *game) {
	switch cev.EAction {
	case CloudEnd:
		delete(g.Clouds, cev.Pos)
		g.ComputeLOS()
	case ObstructionEnd:
		if !g.Player.LOS[cev.Pos] && g.Dungeon.Cell(cev.Pos).T == WallCell {
			g.WrongWall[cev.Pos] = !g.WrongWall[cev.Pos]
		} else {
			delete(g.TemporalWalls, cev.Pos)
		}
		if g.Dungeon.Cell(cev.Pos).T == FreeCell {
			break
		}
		g.Dungeon.SetCell(cev.Pos, FreeCell)
		g.MakeNoise(TemporalWallNoise, cev.Pos)
		g.Fog(cev.Pos, 1, &simpleEvent{ERank: cev.Rank()})
		g.ComputeLOS()
	case ObstructionProgression:
		pos := g.FreeCell()
		g.TemporalWallAt(pos, cev)
		if g.Player.LOS[pos] {
			g.Printf("You see a wall appear out of thin air.")
			g.StopAuto()
		}
		g.PushEvent(&cloudEvent{ERank: cev.Rank() + 200 + RandInt(50), EAction: ObstructionProgression})
	case FireProgression:
		if _, ok := g.Clouds[cev.Pos]; !ok {
			break
		}
		g.BurnCreature(cev.Pos, cev)
		if RandInt(10) == 0 {
			delete(g.Clouds, cev.Pos)
			g.Fog(cev.Pos, 1, &simpleEvent{ERank: cev.Rank()})
			g.ComputeLOS()
			break
		}
		for _, pos := range g.Dungeon.FreeNeighbors(cev.Pos) {
			if RandInt(3) > 0 {
				continue
			}
			g.Burn(pos, cev)
		}
		cev.Renew(g, 10)
	case NightProgression:
		if _, ok := g.Clouds[cev.Pos]; !ok {
			break
		}
		g.MakeCreatureSleep(cev.Pos, cev)
		if RandInt(20) == 0 {
			delete(g.Clouds, cev.Pos)
			g.ComputeLOS()
			break
		}
		cev.Renew(g, 10)
	}
}

func (g *game) MakeCreatureSleep(pos position, ev event) {
	if pos == g.Player.Pos {
		g.Player.Statuses[StatusSlow]++
		g.PushEvent(&simpleEvent{ERank: ev.Rank() + 30 + RandInt(10), EAction: SlowEnd})
		g.Print("The clouds of night make you sleepy.")
		return
	}
	mons := g.MonsterAt(pos)
	if !mons.Exists() || (RandInt(2) == 0 && mons.Status(MonsExhausted)) {
		// do not always make already exhausted monsters sleep (they were probably awaken)
		return
	}
	if mons.State != Resting && g.Player.LOS[mons.Pos] {
		g.Printf("%s falls asleep.", mons.Kind.Definite(true))
	}
	mons.State = Resting
	mons.ExhaustTime(g, 40+RandInt(10))
}

func (g *game) BurnCreature(pos position, ev event) {
	mons := g.MonsterAt(pos)
	if mons.Exists() {
		mons.HP -= 1 + RandInt(10)
		if mons.HP <= 0 {
			if g.Player.LOS[mons.Pos] {
				g.PrintfStyled("%s is killed by the fire.", logPlayerHit, mons.Kind.Definite(true))
			}
			g.HandleKill(mons, ev)
		} else {
			mons.MakeAwareIfHurt(g)
		}
	}
	if pos == g.Player.Pos {
		damage := 1 + RandInt(10)
		if damage > g.Player.HP {
			damage = 1 + RandInt(10)
		}
		g.Player.HP -= damage
		g.PrintfStyled("The fire burns you (%d dmg).", logMonsterHit, damage)
		if g.Player.HP+damage < 10 {
			g.Stats.TimesLucky++
		}
		g.StopAuto()
	}
}

func (g *game) Burn(pos position, ev event) {
	if _, ok := g.Clouds[pos]; ok {
		return
	}
	_, okFungus := g.Fungus[pos]
	_, okDoor := g.Doors[pos]
	if !okFungus && !okDoor {
		return
	}
	g.Stats.Burns++
	foliage := true
	delete(g.Fungus, pos)
	if _, ok := g.Doors[pos]; ok {
		delete(g.Doors, pos)
		foliage = false
		g.Print("The door vanishes in flames.")
	}
	g.Clouds[pos] = CloudFire
	if !g.Player.LOS[pos] {
		if foliage {
			g.WrongFoliage[pos] = true
		} else {
			g.WrongDoor[pos] = true
		}
	} else {
		g.ComputeLOS()
	}
	g.PushEvent(&cloudEvent{ERank: ev.Rank() + 10, EAction: FireProgression, Pos: pos})
	g.BurnCreature(pos, ev)
}

func (cev *cloudEvent) Renew(g *game, delay int) {
	cev.ERank += delay
	g.PushEvent(cev)
}
