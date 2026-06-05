<script lang="ts">
    import {
        Action,
        Biome,
        DraftKind,
        Structure,
        type DraftItem,
    } from "../lib/types";

    export let item: DraftItem | null | undefined;
    export let size: "sm" | "md" | "lg" = "sm";

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
                return "Tower";
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
                return "Expand";
            default:
                return "Action";
        }
    }

    function title(item: DraftItem | null | undefined) {
        if (!item) return "Empty";

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

    function typeLabel(item: DraftItem | null | undefined) {
        if (!item) return "Hand";

        switch (item.Kind) {
            case DraftKind.Tile:
                return "Tile";
            case DraftKind.Upgrade:
                return "Upgrade";
            case DraftKind.Structure:
                return "Build";
            case DraftKind.Action:
                return "Action";
            default:
                return "";
        }
    }

    function icon(item: DraftItem | null | undefined) {
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

    function cardClass(item: DraftItem | null | undefined) {
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
            }
        }

        if (item.Kind === DraftKind.Upgrade) {
            return "border-[#9b7034] bg-[#f2c36b] text-[#142833]";
        }

        if (item.Kind === DraftKind.Structure) {
            return "border-[#6b4a2f] bg-[#ead7aa] text-[#142833]";
        }

        if (item.Kind === DraftKind.Action) {
            return "border-[#327b8d] bg-[#73c4bd] text-[#102b38]";
        }

        return "border-[#6b4a2f] bg-[#ead7aa] text-[#142833]";
    }

    $: sizeClass =
        size === "lg" ? "h-32 p-4" : size === "md" ? "h-24 p-3" : "h-16 p-2";

    $: iconClass =
        size === "lg"
            ? "h-14 w-14 text-3xl rounded-2xl"
            : size === "md"
              ? "h-11 w-11 text-2xl rounded-xl"
              : "h-9 w-9 text-xl rounded-xl";

    $: titleClass =
        size === "lg" ? "text-xl" : size === "md" ? "text-base" : "text-sm";
</script>

<div
    class={[
        "relative overflow-hidden rounded-2xl border-2 shadow-[0_5px_0_rgba(0,0,0,0.18)]",
        sizeClass,
        cardClass(item),
    ].join(" ")}
    title={`${title(item)} ${typeLabel(item)}`}
>
    <div class="flex h-full items-center gap-3">
        <div
            class={[
                "grid shrink-0 place-items-center bg-white/35 font-black",
                iconClass,
            ].join(" ")}
        >
            {icon(item)}
        </div>

        <div class="min-w-0">
            <div
                class="text-[9px] font-black uppercase tracking-wider opacity-70"
            >
                {typeLabel(item)}
            </div>

            <div
                class={["truncate font-black leading-tight", titleClass].join(
                    " ",
                )}
            >
                {title(item)}
            </div>
        </div>
    </div>
</div>
