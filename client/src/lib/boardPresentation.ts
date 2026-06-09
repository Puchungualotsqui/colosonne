import { Biome, Structure, type Tile } from "./types";

export function biomeClass(tile: Tile | undefined, candidate: boolean) {
  if (candidate) {
    return "border-[#d1a45f] bg-[#ead7aa]/35 text-[#6b4a2f]";
  }

  switch (tile?.Biome) {
    case Biome.Forest:
      return "border-[#2f6546] bg-[#5b9368] text-[#17313a]";
    case Biome.Mountain:
      return "border-[#656b73] bg-[#a8adb2] text-[#142833]";
    case Biome.Plain:
      return "border-[#9b7034] bg-[#d9b56a] text-[#142833]";
    case Biome.River:
      return "border-[#327b8d] bg-[#6eb8c5] text-[#102b38]";
    case Biome.Ruins:
      return "border-[#6d4c9b] bg-[#9b79c9] text-[#142833]";
    default:
      return "border-[#6b4a2f] bg-[#ead7aa] text-[#142833]";
  }
}

export function biomeLabel(biome: Biome | undefined) {
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
      return "";
  }
}

export function structureLabel(structure: Structure) {
  switch (structure) {
    case Structure.Outpost:
      return "Outpost";
    case Structure.City:
      return "City";
    case Structure.Settlement:
      return "Settlement";
    case Structure.Bridge:
      return "Bridge";
    case Structure.Watchtower:
      return "Watchtower";
    default:
      return "";
  }
}
