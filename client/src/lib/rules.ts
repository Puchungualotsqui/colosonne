import {
  Action,
  Biome,
  DraftKind,
  Structure,
  type DraftItem,
  type GameState,
  type TargetBuildAction,
  type Tile,
} from "./types";

export type Coord = {
  x: number;
  y: number;
};

export function tileKey(x: number, y: number) {
  return `${x},${y}`;
}

export function hexNeighbors(x: number, y: number): Coord[] {
  return [
    { x: x + 1, y },
    { x: x + 1, y: y - 1 },
    { x, y: y - 1 },
    { x: x - 1, y },
    { x: x - 1, y: y + 1 },
    { x, y: y + 1 },
  ];
}

export function tileAt(game: GameState, x: number, y: number) {
  return game.Map.find((tile) => tile.X === x && tile.Y === y);
}

export function controlsTile(tile: Tile | undefined, playerId: number) {
  return !!tile && tile.HasOwner && tile.Owner === playerId;
}

export function isEnemyControlledTile(
  tile: Tile | undefined,
  playerId: number,
) {
  return !!tile && tile.HasOwner && tile.Owner !== playerId;
}

export function isNeutralTile(tile: Tile | undefined) {
  return !!tile && !tile.HasOwner;
}

export function hasAdjacentControlledTile(
  game: GameState,
  x: number,
  y: number,
  playerId: number,
) {
  return hexNeighbors(x, y).some((coord) =>
    controlsTile(tileAt(game, coord.x, coord.y), playerId),
  );
}

export function isUnbridgedRiver(tile: Tile | undefined) {
  return (
    !!tile && tile.Biome === Biome.River && tile.Structure !== Structure.Bridge
  );
}

export function canReceiveInfluence(tile: Tile | undefined) {
  if (!tile) return false;
  if (isUnbridgedRiver(tile)) return false;
  return true;
}

export function isActiveStructure(tile: Tile | undefined) {
  if (!tile) return false;
  if (tile.Structure === Structure.None) return false;
  if (tile.HasBlockade) return false;

  return tile.HasOwner && tile.Owner === tile.StructureOwner;
}

export function tileHasProducingBiome(tile: Tile | undefined) {
  if (!tile) return false;

  return (
    tile.Biome === Biome.Forest ||
    tile.Biome === Biome.Mountain ||
    tile.Biome === Biome.Plain ||
    tile.Biome === Biome.Ruins
  );
}

export function canHarvestTile(
  game: GameState,
  playerId: number,
  tile: Tile | undefined,
) {
  if (!tile) return false;
  if (!controlsTile(tile, playerId)) return false;
  if (!isActiveStructure(tile)) return false;
  if (!tileHasProducingBiome(tile)) return false;

  return (
    tile.Structure === Structure.Settlement ||
    tile.Structure === Structure.Outpost ||
    tile.Structure === Structure.City
  );
}

export function canBuildOnTile(
  game: GameState,
  playerId: number,
  action: TargetBuildAction,
  tile: Tile | undefined,
) {
  if (!tile) return false;

  switch (action) {
    case "outpost":
      return (
        tile.Biome !== Biome.River &&
        tile.Structure === Structure.None &&
        !tile.HasBlockade &&
        !isEnemyControlledTile(tile, playerId)
      );

    case "settlement":
      return (
        tile.Biome !== Biome.River &&
        tile.Structure === Structure.None &&
        !tile.HasBlockade &&
        controlsTile(tile, playerId)
      );

    case "city":
      return (
        tile.Structure === Structure.Outpost &&
        tile.StructureOwner === playerId &&
        !tile.HasBlockade &&
        controlsTile(tile, playerId)
      );

    case "blockade":
      return (
        tile.Biome !== Biome.River &&
        !tile.HasBlockade &&
        !controlsTile(tile, playerId) &&
        hasAdjacentControlledTile(game, tile.X, tile.Y, playerId)
      );

    case "flood":
      return tile.Biome !== Biome.River && tile.Structure === Structure.None;

    default:
      return false;
  }
}

export function canUseDraftOnTile(
  game: GameState,
  playerId: number,
  item: DraftItem | null | undefined,
  tile: Tile | undefined,
) {
  if (!item || !tile) return false;

  if (item.Kind === DraftKind.Structure) {
    if (tile.Structure !== Structure.None) return false;

    switch (item.Structure) {
      case Structure.Bridge:
        return tile.Biome === Biome.River;

      case Structure.Watchtower:
        return (
          tile.Biome !== Biome.River &&
          !tile.HasBlockade &&
          !isEnemyControlledTile(tile, playerId)
        );

      default:
        return false;
    }
  }

  if (item.Kind === DraftKind.Action) {
    switch (item.Action) {
      case Action.Harvest:
        return canHarvestTile(game, playerId, tile);

      case Action.Reinforce:
        return canReceiveInfluence(tile);

      default:
        return false;
    }
  }

  return false;
}

export function draftNeedsBoardTarget(item: DraftItem | null | undefined) {
  if (!item) return false;
  if (item.Kind === DraftKind.Structure) return true;

  return (
    item.Kind === DraftKind.Action &&
    (item.Action === Action.Harvest || item.Action === Action.Reinforce)
  );
}

export function canUseHandItem(
  game: GameState,
  playerId: number,
  item: DraftItem | null | undefined,
) {
  if (!item) return false;

  switch (item.Kind) {
    case DraftKind.Tile:
      return game.Map.some((tile) =>
        hexNeighbors(tile.X, tile.Y).some(
          (coord) => !tileAt(game, coord.x, coord.y),
        ),
      );

    case DraftKind.Structure:
    case DraftKind.Action:
      if (item.Kind === DraftKind.Action) {
        if (item.Action === Action.Expansion) return true;
        if (item.Action === Action.Raid) {
          return game.Players.some((player) => player.Id !== playerId);
        }
      }

      return game.Map.some((tile) =>
        canUseDraftOnTile(game, playerId, item, tile),
      );

    default:
      return false;
  }
}

export function hasBuildTarget(
  game: GameState,
  playerId: number,
  action: TargetBuildAction,
) {
  return game.Map.some((tile) => canBuildOnTile(game, playerId, action, tile));
}
