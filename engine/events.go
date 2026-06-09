package engine

type GameEventKind string

const (
	EventTilePlaced        GameEventKind = "tile_placed"
	EventStructurePlaced   GameEventKind = "structure_placed"
	EventStructureUpgraded GameEventKind = "structure_upgraded"
	EventBlockadePlaced    GameEventKind = "blockade_placed"
	EventFloodTile         GameEventKind = "flood_tile"

	EventResourceGain     GameEventKind = "resource_gain"
	EventResourceTransfer GameEventKind = "resource_transfer"
	EventInfluenceAdded   GameEventKind = "influence_added"
	EventActionUsed       GameEventKind = "action_used"
	EventFloodworksBought GameEventKind = "floodworks_bought"
)

type EventCoord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type GameEvent struct {
	ID   uint64        `json:"id"`
	Kind GameEventKind `json:"kind"`

	Actor PlayerId `json:"actor,omitempty"`

	FromPlayer PlayerId `json:"fromPlayer,omitempty"`
	ToPlayer   PlayerId `json:"toPlayer,omitempty"`

	From *EventCoord `json:"from,omitempty"`
	To   *EventCoord `json:"to,omitempty"`

	Resource Resource `json:"resource,omitempty"`
	Amount   uint     `json:"amount,omitempty"`

	Biome     Biome     `json:"biome,omitempty"`
	Structure Structure `json:"structure,omitempty"`
	Action    Action    `json:"action,omitempty"`
}

func eventCoord(x, y int) *EventCoord {
	return &EventCoord{X: x, Y: y}
}

func (gs *GameState) emitEvent(event GameEvent) {
	if gs == nil {
		return
	}

	gs.eventSeq++
	event.ID = gs.eventSeq

	gs.Events = append(gs.Events, event)
}

func (gs *GameState) DrainEvents() []GameEvent {
	if gs == nil || len(gs.Events) == 0 {
		return []GameEvent{}
	}

	events := gs.Events
	gs.Events = nil
	return events
}

func (gs *GameState) EmitResourceGainFromTile(
	playerId PlayerId,
	resource Resource,
	amount uint,
	x int,
	y int,
	action Action,
) {
	if resource == NoneResource || amount == 0 {
		return
	}

	gs.emitEvent(GameEvent{
		Kind:     EventResourceGain,
		Actor:    playerId,
		ToPlayer: playerId,
		From:     eventCoord(x, y),
		Resource: resource,
		Amount:   amount,
		Action:   action,
	})
}

func (gs *GameState) EmitResourceGainFromAction(
	playerId PlayerId,
	resource Resource,
	amount uint,
	action Action,
) {
	if resource == NoneResource || amount == 0 {
		return
	}

	gs.emitEvent(GameEvent{
		Kind:     EventResourceGain,
		Actor:    playerId,
		ToPlayer: playerId,
		Resource: resource,
		Amount:   amount,
		Action:   action,
	})
}

func (gs *GameState) EmitResourceTransfer(
	fromPlayer PlayerId,
	toPlayer PlayerId,
	resource Resource,
	amount uint,
	action Action,
) {
	if resource == NoneResource || amount == 0 {
		return
	}

	gs.emitEvent(GameEvent{
		Kind:       EventResourceTransfer,
		Actor:      toPlayer,
		FromPlayer: fromPlayer,
		ToPlayer:   toPlayer,
		Resource:   resource,
		Amount:     amount,
		Action:     action,
	})
}
