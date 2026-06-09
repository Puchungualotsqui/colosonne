export enum GamePhase {
  Pick = 0,
  Place = 1,
  Build = 2,
}

export enum DraftKind {
  Tile = 0,
  Structure = 1,
  Action = 2,
}

export enum Resource {
  None = 0,
  Wood = 1,
  Stone = 2,
  Grain = 3,
  Relic = 4,
}

export enum Biome {
  None = 0,
  Forest = 1,
  Mountain = 2,
  Plain = 3,
  River = 4,
  Ruins = 5,
}

export enum Structure {
  None = 0,
  Outpost = 1,
  City = 2,
  Settlement = 3,
  Bridge = 4,
  Watchtower = 5,
}

export enum Action {
  None = 0,
  Harvest = 1,
  Reinforce = 2,
  Expansion = 3,
  Raid = 4,
}

export type BuildAction =
  | "outpost"
  | "settlement"
  | "city"
  | "blockade"
  | "floodworks"
  | "flood"
  | "pass";

export type TargetBuildAction =
  | "outpost"
  | "settlement"
  | "city"
  | "blockade"
  | "flood";

export type DraftItem = {
  Kind: DraftKind;
  Biome: Biome;
  Structure: Structure;
  Action: Action;
};

export type MarketSlot = DraftItem | null;

export type Player = {
  Id: number;
  Hand: DraftItem[] | null;
  Resources: Record<number, number>;

  FloodTokens?: number;

  FloodworksBought?: number;
};

export type ResourceCostResponse = {
  wood: number;
  stone: number;
  grain: number;
  relic: number;
};

export type BuildCostsResponse = {
  outpost: ResourceCostResponse;
  city: ResourceCostResponse;
  settlement: ResourceCostResponse;
  blockade: ResourceCostResponse;
  floodworks: ResourceCostResponse;
};

export type BuildCostsByPlayer = Record<string, BuildCostsResponse>;

export type Tile = {
  X: number;
  Y: number;

  Biome: Biome;

  Influence: Record<number, number>;
  TempInfluence: Record<number, number>;

  Owner: number;
  HasOwner: boolean;

  Structure: Structure;
  StructureOwner: number;

  HasBlockade: boolean;
  BlockadeOwner: number;
};

export type GameState = {
  Players: Player[];
  Map: Tile[];
  Deck: DraftItem[];

  // Backend sends 6 fixed slots.
  // Picked slots are null.
  Market: MarketSlot[];

  CurrentPlayer: number;
  CurrentPhase: GamePhase;
  RoundFirstIndex: number;
  TurnIndex: number;
  Round: number;
};

export type RoomIdentity = {
  roomId: string;
  playerId: number;
  name: string;
  isGuest: boolean;
  role: "player" | "spectator";
};

export type RoomPlayer = {
  playerId: number;
  clientId: string;
  userId?: number;
  name: string;
  isGuest: boolean;
  ready: boolean;
  isHost: boolean;
};

export type RoomSpectator = {
  clientId: string;
  userId?: number;
  name: string;
  isGuest: boolean;
};

export type RoomSettings = {
  maxPlayers: number;
  spectators: boolean;
};

export type RoomState = {
  roomId: string;
  status: "lobby" | "playing" | "ended";
  settings: RoomSettings;
  players: RoomPlayer[];
  spectators: RoomSpectator[];
  game: GameState | null;

  buildCosts?: BuildCostsByPlayer;

  // Transient backend events used for animation.
  events?: GameEvent[];
};

export type ServerMessage =
  | { type: "room_created"; data: RoomIdentity }
  | { type: "room_joined"; data: RoomIdentity }
  | { type: "room_spectating"; data: RoomIdentity }
  | { type: "room_state"; data: RoomState }
  | {
      type: "room_waiting";
      data: {
        roomId: string;
        players: number;
        spectators: number;
      };
    }
  | {
      type: "state";
      data: {
        roomId: string;
        players: number;
        spectators: number;
        game: GameState;
      };
    }
  | { type: "kicked"; data: string }
  | { type: "error"; data: string };

export type EventCoord = {
  x: number;
  y: number;
};

export type GameEventKind =
  | "tile_placed"
  | "structure_placed"
  | "structure_upgraded"
  | "blockade_placed"
  | "flood_tile"
  | "resource_gain"
  | "resource_transfer"
  | "influence_added"
  | "action_used"
  | "floodworks_bought";

export type GameEvent = {
  id: number;
  kind: GameEventKind;

  actor?: number;

  fromPlayer?: number;
  toPlayer?: number;

  from?: EventCoord;
  to?: EventCoord;

  resource?: Resource;
  amount?: number;

  biome?: Biome;
  structure?: Structure;
  action?: Action;
};
