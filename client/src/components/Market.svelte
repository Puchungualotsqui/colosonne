<script lang="ts">
    import {
        Action,
        Biome,
        DraftKind,
        GamePhase,
        Structure,
        type DraftItem,
        type GameState,
    } from "../lib/types";

    export let game: GameState;
    export let playerId = 0;
    export let role: "player" | "spectator" | "" = "";

    export let onPick: (marketIndex: number) => void;

    $: canPick =
        role === "player" &&
        game.CurrentPhase === GamePhase.Pick &&
        game.CurrentPlayer === playerId;

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

    function itemTitle(item: DraftItem) {
        switch (item.Kind) {
            case DraftKind.Tile:
                return `${biomeName(item.Biome)} Tile`;
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

    function itemDescription(item: DraftItem) {
        switch (item.Kind) {
            case DraftKind.Tile:
                return "Place this hex on the frontier.";
            case DraftKind.Upgrade:
                return "Improve a controlled non-river tile.";
            case DraftKind.Structure:
                if (item.Structure === Structure.Bridge) {
                    return "Build on a river adjacent to your territory.";
                }
                return "Build on a controlled land tile.";
            case DraftKind.Action:
                if (item.Action === Action.Harvest) {
                    return "Gain extra resources from a controlled tile.";
                }
                if (item.Action === Action.Reinforce) {
                    return "Add temporary influence to a tile.";
                }
                if (item.Action === Action.Expansion) {
                    return "Gain Wood and Grain.";
                }
                return "Use a special action.";
            default:
                return "";
        }
    }

    function itemClass(item: DraftItem) {
        if (item.Kind === DraftKind.Tile) {
            switch (item.Biome) {
                case Biome.Forest:
                    return "bg-[#5b9368] border-[#2f6546]";
                case Biome.Mountain:
                    return "bg-[#a8adb2] border-[#656b73]";
                case Biome.Plain:
                    return "bg-[#d9b56a] border-[#9b7034]";
                case Biome.River:
                    return "bg-[#6eb8c5] border-[#327b8d]";
            }
        }

        if (item.Kind === DraftKind.Upgrade) {
            return "bg-[#f2c36b] border-[#9b7034]";
        }

        if (item.Kind === DraftKind.Structure) {
            return "bg-[#ead7aa] border-[#6b4a2f]";
        }

        if (item.Kind === DraftKind.Action) {
            return "bg-[#73c4bd] border-[#327b8d]";
        }

        return "bg-[#ead7aa] border-[#6b4a2f]";
    }
</script>

<section
    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
>
    <div class="flex items-center justify-between gap-3">
        <div>
            <h2 class="text-xl font-black text-[#fff7e8]">Market</h2>
            <p class="mt-1 text-sm font-semibold text-[#9fc9c5]">
                {canPick ? "Choose one draft card." : "Drafting is locked."}
            </p>
        </div>

        <div
            class="rounded-xl bg-[#f8efe0]/10 px-3 py-1 text-sm font-bold text-[#9fc9c5]"
        >
            {game.Market.length} cards
        </div>
    </div>

    <div class="mt-4 grid gap-3">
        {#each game.Market as item, index}
            <button
                class={[
                    "cursor-pointer rounded-2xl border-2 p-4 text-left text-[#142833] shadow-[0_6px_0_rgba(0,0,0,0.18)] transition hover:-translate-y-0.5 active:translate-y-1 disabled:cursor-not-allowed disabled:opacity-55 disabled:hover:translate-y-0",
                    itemClass(item),
                ].join(" ")}
                type="button"
                disabled={!canPick}
                on:click={() => onPick(index)}
            >
                <div class="flex items-start justify-between gap-3">
                    <div>
                        <div class="text-lg font-black">
                            {itemTitle(item)}
                        </div>

                        <div class="mt-1 text-sm font-semibold opacity-80">
                            {itemDescription(item)}
                        </div>
                    </div>

                    <div
                        class="rounded-xl bg-white/35 px-2 py-1 text-xs font-black"
                    >
                        #{index}
                    </div>
                </div>
            </button>
        {/each}
    </div>
</section>
