<script lang="ts">
    import {
        Action,
        Biome,
        DraftKind,
        Structure,
        type DraftItem,
    } from "../lib/types";

    export let item: DraftItem;
    export let index = 0;
    export let disabled = false;
    export let selected = false;
    export let onPick: (index: number) => void;

    function biomeName(biome: Biome) {
        switch (biome) {
            case Biome.Forest:
                return "Forest";
            case Biome.Mountain:
                return "Mountain";
            case Biome.Plain:
                return "Plain";
            case Biome.River:
                return "River";
            default:
                return "Unknown";
        }
    }

    function structureName(structure: Structure) {
        switch (structure) {
            case Structure.Bridge:
                return "Bridge";
            case Structure.Watchtower:
                return "Watchtower";
            case Structure.Road:
                return "Road";
            case Structure.Outpost:
                return "Outpost";
            case Structure.City:
                return "City";
            default:
                return "Structure";
        }
    }

    function actionName(action: Action) {
        switch (action) {
            case Action.Harvest:
                return "Harvest";
            case Action.Reinforce:
                return "Reinforce";
            case Action.Expansion:
                return "Expansion";
            default:
                return "Action";
        }
    }

    function title(item: DraftItem) {
        switch (item.Kind) {
            case DraftKind.Tile:
                return biomeName(item.Biome);
            case DraftKind.Upgrade:
                return "Upgrade";
            case DraftKind.Structure:
                return structureName(item.Structure);
            case DraftKind.Action:
                return actionName(item.Action);
            default:
                return "Unknown";
        }
    }

    function subtitle(item: DraftItem) {
        switch (item.Kind) {
            case DraftKind.Tile:
                return "Tile";
            case DraftKind.Upgrade:
                return "Improve tile";
            case DraftKind.Structure:
                return "Structure";
            case DraftKind.Action:
                return "Action";
            default:
                return "";
        }
    }

    function icon(item: DraftItem) {
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
            }
        }

        if (item.Kind === DraftKind.Upgrade) return "↑";

        if (item.Kind === DraftKind.Structure) {
            switch (item.Structure) {
                case Structure.Outpost:
                    return "⌂";
                case Structure.City:
                    return "▦";
                case Structure.Road:
                    return "━";
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
            }
        }

        return "?";
    }

    function description(item: DraftItem) {
        switch (item.Kind) {
            case DraftKind.Tile:
                return "Place this hex adjacent to the existing map.";
            case DraftKind.Upgrade:
                return "Upgrade one controlled non-river tile. Higher levels produce more.";
            case DraftKind.Structure:
                if (item.Structure === Structure.Bridge) {
                    return "Build on a river adjacent to your territory.";
                }
                if (item.Structure === Structure.Watchtower) {
                    return "Project influence up to distance 2.";
                }
                if (item.Structure === Structure.Road) {
                    return "Connect territory and compete for longest road.";
                }
                return "Place a structure on a valid controlled land tile.";
            case DraftKind.Action:
                if (item.Action === Action.Harvest) {
                    return "Gain extra resources from one controlled non-river tile.";
                }
                if (item.Action === Action.Reinforce) {
                    return "Add temporary influence to one tile for the next recount.";
                }
                if (item.Action === Action.Expansion) {
                    return "Gain 1 Wood and 1 Grain immediately.";
                }
                return "Use this action during your place/use step.";
            default:
                return "";
        }
    }

    function cardClass(item: DraftItem) {
        if (item.Kind === DraftKind.Tile) {
            switch (item.Biome) {
                case Biome.Forest:
                    return "border-[#2f6546] bg-[#5b9368]";
                case Biome.Mountain:
                    return "border-[#656b73] bg-[#a8adb2]";
                case Biome.Plain:
                    return "border-[#9b7034] bg-[#d9b56a]";
                case Biome.River:
                    return "border-[#327b8d] bg-[#6eb8c5]";
            }
        }

        if (item.Kind === DraftKind.Upgrade) {
            return "border-[#9b7034] bg-[#f2c36b]";
        }

        if (item.Kind === DraftKind.Structure) {
            return "border-[#6b4a2f] bg-[#ead7aa]";
        }

        if (item.Kind === DraftKind.Action) {
            return "border-[#327b8d] bg-[#73c4bd]";
        }

        return "border-[#6b4a2f] bg-[#ead7aa]";
    }
</script>

<button
    class={[
        "group relative h-32 rounded-2xl border-2 p-3 text-left text-[#142833] shadow-[0_6px_0_rgba(0,0,0,0.18)] transition",
        "hover:-translate-y-1 hover:brightness-105 active:translate-y-1",
        "disabled:cursor-not-allowed disabled:opacity-55 disabled:hover:translate-y-0",
        selected ? "ring-4 ring-[#f2c36b]" : "",
        cardClass(item),
    ].join(" ")}
    type="button"
    {disabled}
    on:click={() => onPick(index)}
>
    <div class="flex h-full flex-col justify-between">
        <div class="flex items-start justify-between gap-2">
            <div
                class="grid h-10 w-10 place-items-center rounded-xl bg-white/35 text-2xl font-black"
            >
                {icon(item)}
            </div>

            <div
                class="rounded-lg bg-white/30 px-2 py-1 text-[10px] font-black uppercase tracking-wider"
            >
                {subtitle(item)}
            </div>
        </div>

        <div>
            <div class="text-lg font-black leading-5">
                {title(item)}
            </div>
        </div>
    </div>

    <div
        class="pointer-events-none absolute bottom-[calc(100%+10px)] left-1/2 z-50 hidden w-64 -translate-x-1/2 rounded-2xl bg-[#142833] p-4 text-[#f8efe0] shadow-xl ring-1 ring-white/10 group-hover:block"
    >
        <div class="text-sm font-black">
            {title(item)}
        </div>
        <div class="mt-1 text-xs font-semibold leading-5 text-[#d9e6df]">
            {description(item)}
        </div>
    </div>
</button>
