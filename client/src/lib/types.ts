export enum GamePhase {
  Pick = 0,
  Place = 1,
  Build = 2,
  Recount = 3,
}

export enum DraftKind {
  Tile = 0,
  Upgrade = 1,
  Structure = 2,
  Action = 3,
}

export enum Biome {
  None = 0,
  Forest = 1,
  Mountain = 2,
  Plain = 3,
  River = 4,
}

export enum Structure {
  None = 0,
  Outpost = 1,
  City = 2,
  Bridge = 3,
  Watchtower = 4,
  Road = 5,
}

export enum Action {
  None = 0,
  Harvest = 1,
  Reinforce = 2,
  Expansion = 3,
}

export type DraftItem = {
  Kind: DraftKind;
  Biome: Biome;
  Structure: Structure;
  Action: Action;
};

export type Player = {
  Id: number;
  Hand: DraftItem | null;
  Resources: Record<number, number>;
};

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
  UpgradeLevel: number;
};

export type GameState = {
  Players: Player[];
  Map: Tile[];
  Deck: DraftItem[];
  Market: DraftItem[];
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
