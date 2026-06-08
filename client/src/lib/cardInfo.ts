import { Action, Biome, DraftKind, Structure, type DraftItem } from "./types";

export function biomeName(biome: Biome) {
  switch (biome) {
    case Biome.Forest:
      return "Forest";
    case Biome.Mountain:
      return "Mountain";
    case Biome.Plain:
      return "Plain";
    case Biome.River:
      return "River";
    case Biome.Ruins:
      return "Ruins";
    default:
      return "Unknown";
  }
}

export function structureName(structure: Structure) {
  switch (structure) {
    case Structure.Bridge:
      return "Bridge";
    case Structure.Watchtower:
      return "Watchtower";
    case Structure.Outpost:
      return "Outpost";
    case Structure.City:
      return "City";
    case Structure.Settlement:
      return "Settlement";
    default:
      return "Structure";
  }
}

export function actionName(action: Action) {
  switch (action) {
    case Action.Harvest:
      return "Harvest";
    case Action.Reinforce:
      return "Reinforce";
    case Action.Expansion:
      return "Expansion";
    case Action.Raid:
      return "Raid";
    default:
      return "Action";
  }
}

export function cardTitle(item: DraftItem | null | undefined) {
  if (!item) return "Empty";

  switch (item.Kind) {
    case DraftKind.Tile:
      return biomeName(item.Biome);
    case DraftKind.Structure:
      return structureName(item.Structure);
    case DraftKind.Action:
      return actionName(item.Action);
    default:
      return "Unknown";
  }
}

export function cardType(item: DraftItem | null | undefined) {
  if (!item) return "Card";

  switch (item.Kind) {
    case DraftKind.Tile:
      return "Tile";
    case DraftKind.Structure:
      return "Structure";
    case DraftKind.Action:
      return "Action";
    default:
      return "";
  }
}

export function cardIcon(item: DraftItem | null | undefined) {
  if (!item) return "—";

  if (item.Kind === DraftKind.Tile) {
    switch (item.Biome) {
      case Biome.Forest:
        return "♣";
      case Biome.Mountain:
        return "▲";
      case Biome.Plain:
        return "◆";
      case Biome.River:
        return "≈";
      case Biome.Ruins:
        return "✧";
    }
  }

  if (item.Kind === DraftKind.Structure) {
    switch (item.Structure) {
      case Structure.Outpost:
        return "⌂";
      case Structure.City:
        return "▦";
      case Structure.Settlement:
        return "◈";
      case Structure.Bridge:
        return "⌒";
      case Structure.Watchtower:
        return "♜";
    }
  }

  if (item.Kind === DraftKind.Action) {
    switch (item.Action) {
      case Action.Harvest:
        return "✦";
      case Action.Reinforce:
        return "+";
      case Action.Expansion:
        return "⇱";
      case Action.Raid:
        return "☠";
    }
  }

  return "?";
}

export function cardDescription(item: DraftItem | null | undefined) {
  if (!item) return "";

  switch (item.Kind) {
    case DraftKind.Tile:
      switch (item.Biome) {
        case Biome.Forest:
          return "Place a Forest tile. Produces Wood when controlled with an active structure.";
        case Biome.Mountain:
          return "Place a Mountain tile. Produces Stone when controlled with an active structure.";
        case Biome.Plain:
          return "Place a Plain tile. Produces Grain and supports cities.";
        case Biome.River:
          return "Place a River tile. Rivers block normal control unless bridged.";
        case Biome.Ruins:
          return "Place Ruins. Produces Relic when controlled with an active structure.";
        default:
          return "Place this tile adjacent to the existing map.";
      }

    case DraftKind.Structure:
      if (item.Structure === Structure.Bridge) {
        return "Draft-only structure. Place on a River adjacent to your controlled territory.";
      }

      if (item.Structure === Structure.Watchtower) {
        return "Draft-only structure. Place on your land or neutral non-river land. Strong influence, no production.";
      }

      return "This structure is not normally played from draft.";

    case DraftKind.Action:
      if (item.Action === Action.Harvest) {
        return "Choose one controlled tile with an active structure and gain 2 of its resource.";
      }

      if (item.Action === Action.Reinforce) {
        return "Add +2 temporary influence to one valid tile.";
      }

      if (item.Action === Action.Expansion) {
        return "Gain 1 Wood and 1 Grain immediately.";
      }

      if (item.Action === Action.Raid) {
        return "Steal up to 3 random resources from another player.";
      }

      return "Use this action during your Use phase.";

    default:
      return "";
  }
}

export function cardClass(item: DraftItem | null | undefined) {
  if (!item) {
    return "border-[#f8efe0]/15 bg-[#f8efe0]/8 text-[#9fc9c5]";
  }

  if (item.Kind === DraftKind.Tile) {
    switch (item.Biome) {
      case Biome.Forest:
        return "border-[#2f6546] bg-[#5b9368] text-[#142833]";
      case Biome.Mountain:
        return "border-[#656b73] bg-[#a8adb2] text-[#142833]";
      case Biome.Plain:
        return "border-[#9b7034] bg-[#d9b56a] text-[#142833]";
      case Biome.River:
        return "border-[#327b8d] bg-[#6eb8c5] text-[#102b38]";
      case Biome.Ruins:
        return "border-[#6d4c9b] bg-[#9b79c9] text-[#142833]";
    }
  }

  if (item.Kind === DraftKind.Structure) {
    return "border-[#6b4a2f] bg-[#ead7aa] text-[#142833]";
  }

  if (item.Kind === DraftKind.Action) {
    return "border-[#327b8d] bg-[#73c4bd] text-[#102b38]";
  }

  return "border-[#6b4a2f] bg-[#ead7aa] text-[#142833]";
}
