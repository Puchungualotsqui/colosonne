<script lang="ts">
    import Board from "./Board.svelte";
    import Market from "./Market.svelte";
    import {
        Action,
        Biome,
        DraftKind,
        GamePhase,
        Structure,
        type DraftItem,
        type GameState,
        type Player,
    } from "../lib/types";
    import { debugLog } from "../lib/debug";
    import HandCard from "./HandCard.svelte";

    export let game: GameState;
    export let roomId = "";
    export let playerId = 0;
    export let role: "player" | "spectator" | "" = "";
    export let error = "";

    export let onPick: (marketIndex: number) => void;
    export let onPlaceTile: (x: number, y: number) => void;
    export let onUseDraft: (x: number, y: number) => void;
    export let onPassPlace: () => void;
    export let onBuild: (
        action: "outpost" | "city",
        x: number,
        y: number,
    ) => void;
    export let onPassBuild: () => void;
    export let onLeaveRoom: () => void;
    export let onCopyRoomCode: () => void;

    let selectedBuildAction: "outpost" | "city" | null = null;

    $: me = game.Players.find((p) => p.Id === playerId);
    $: isMyTurn = role === "player" && game.CurrentPlayer === playerId;
    $: currentPhaseName = phaseName(game.CurrentPhase);
    $: currentPlayerName =
        game.CurrentPlayer === playerId
            ? "You"
            : `Player ${game.CurrentPlayer}`;

    $: if (game.CurrentPhase !== GamePhase.Build) {
        selectedBuildAction = null;
    }

    $: debugLog("gameview.state", {
        roomId,
        role,
        playerId,
        currentPlayer: game.CurrentPlayer,
        currentPhase: game.CurrentPhase,
        round: game.Round,
        isMyTurn,
        me,
        selectedBuildAction,
    });

    $: handIsUsable = canUseHandLocally(me?.Hand);
    $: canPassPlace =
        isMyTurn && game.CurrentPhase === GamePhase.Place && !handIsUsable;

    function phaseName(phase: GamePhase) {
        switch (phase) {
            case GamePhase.Pick:
                return "Draft";
            case GamePhase.Place:
                return "Place";
            case GamePhase.Build:
                return "Build";
            default:
                return "Unknown";
        }
    }

    function resourceAmount(player: Player, resourceId: number) {
        return player.Resources?.[resourceId] ?? 0;
    }

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

    function draftName(item: DraftItem | null | undefined) {
        if (!item) return "Empty";

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

    function playerColor(playerId: number) {
        if (playerId === 1) return "bg-[#1d4e89]";
        if (playerId === 2) return "bg-[#b94b3f]";
        return "bg-[#6b4a2f]";
    }

    function handleBuild(action: "outpost" | "city", x: number, y: number) {
        debugLog("build.send", {
            action,
            x,
            y,
            playerId,
            currentPlayer: game.CurrentPlayer,
            currentPhase: game.CurrentPhase,
        });

        onBuild(action, x, y);
        selectedBuildAction = null;
    }

    function hexNeighbors(x: number, y: number) {
        return [
            { x: x + 1, y },
            { x: x + 1, y: y - 1 },
            { x, y: y - 1 },
            { x: x - 1, y },
            { x: x - 1, y: y + 1 },
            { x, y: y + 1 },
        ];
    }

    function tileAt(x: number, y: number) {
        return game.Map.find((t) => t.X === x && t.Y === y);
    }

    function controlsTile(
        tile: { HasOwner: boolean; Owner: number } | undefined,
    ) {
        return !!tile && tile.HasOwner && tile.Owner === playerId;
    }

    function hasAdjacentControlledTile(x: number, y: number) {
        return hexNeighbors(x, y).some((n) => controlsTile(tileAt(n.x, n.y)));
    }

    function canUseHandLocally(item: DraftItem | null | undefined) {
        if (!item) return false;

        switch (item.Kind) {
            case DraftKind.Tile:
                return game.Map.some((tile) =>
                    hexNeighbors(tile.X, tile.Y).some((n) => !tileAt(n.x, n.y)),
                );

            case DraftKind.Upgrade:
                return game.Map.some(
                    (tile) =>
                        tile.Biome !== Biome.River &&
                        controlsTile(tile) &&
                        tile.UpgradeLevel < 3,
                );

            case DraftKind.Structure:
                return game.Map.some((tile) =>
                    canUseStructureOnTileLocally(item.Structure, tile),
                );

            case DraftKind.Action:
                if (item.Action === Action.Expansion) return true;
                if (item.Action === Action.Reinforce)
                    return game.Map.length > 0;

                if (item.Action === Action.Harvest) {
                    return game.Map.some(
                        (tile) =>
                            controlsTile(tile) && tile.Biome !== Biome.River,
                    );
                }

                return false;

            default:
                return false;
        }
    }

    function canUseStructureOnTileLocally(structure: Structure, tile: any) {
        if (!tile) return false;

        if (tile.Structure !== Structure.None) {
            return false;
        }

        switch (structure) {
            case Structure.Bridge:
                return (
                    tile.Biome === Biome.River &&
                    hasAdjacentControlledTile(tile.X, tile.Y)
                );

            case Structure.Road:
            case Structure.Watchtower:
                return tile.Biome !== Biome.River && controlsTile(tile);

            case Structure.Outpost:
                return (
                    tile.Biome !== Biome.River &&
                    (!tile.HasOwner || tile.Owner === playerId)
                );

            case Structure.City:
                return false;

            default:
                return false;
        }
    }
</script>

<main
    class="relative min-h-screen overflow-hidden bg-[#17313a] font-sans text-[#f8efe0]"
>
    <div class="pointer-events-none absolute inset-0 bg-[#15323a]">
        <div class="absolute inset-0 bg-board-flat"></div>
        <div class="absolute inset-0 bg-board-texture opacity-[0.16]"></div>
        <div
            class="absolute inset-x-0 bottom-0 h-40 bg-gradient-to-t from-black/18 to-transparent"
        ></div>
    </div>

    <header
        class="relative z-10 mx-auto flex max-w-7xl items-center justify-between px-6 py-5 lg:px-12"
    >
        <div class="flex items-center gap-3">
            <div
                class="grid h-11 w-11 place-items-center rounded-2xl bg-[#f2c36b] text-xl font-black text-[#142833] shadow-[0_8px_0_rgba(0,0,0,0.16)] ring-1 ring-white/20"
            >
                <span class="logo-diamond">◈</span>
            </div>

            <div>
                <div
                    class="text-xl font-semibold tracking-tight text-[#fff7e8]"
                >
                    Frontiers
                </div>
                <div
                    class="text-xs font-semibold uppercase tracking-[0.22em] text-[#9fc9c5]"
                >
                    Room {roomId}
                </div>
            </div>
        </div>

        <div class="flex items-center gap-3">
            <button
                class="cursor-pointer rounded-xl bg-[#f8efe0]/10 px-3 py-2 text-sm font-bold text-[#fff7e8] ring-1 ring-[#f8efe0]/20 hover:bg-[#f8efe0]/16"
                type="button"
                on:click={onCopyRoomCode}
            >
                Copy Code
            </button>

            <button
                class="cursor-pointer rounded-xl bg-[#b94b3f] px-4 py-2 text-sm font-bold text-white shadow-sm transition hover:bg-[#c9574a]"
                type="button"
                on:click={onLeaveRoom}
            >
                Leave
            </button>
        </div>
    </header>

    <section
        class="relative z-10 mx-auto grid max-w-7xl gap-6 px-6 pb-12 pt-4 lg:grid-cols-[280px_1fr_340px] lg:px-12"
    >
        <aside class="space-y-5">
            <section
                class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
            >
                <div
                    class="text-sm font-black uppercase tracking-[0.22em] text-[#9fc9c5]"
                >
                    Match
                </div>

                <div class="mt-4 grid grid-cols-3 gap-2">
                    <div
                        class="rounded-2xl bg-[#f8efe0]/10 p-3 text-center ring-1 ring-[#f8efe0]/10"
                    >
                        <div
                            class="text-[10px] font-black uppercase tracking-wider text-[#9fc9c5]"
                        >
                            Round
                        </div>
                        <div class="mt-1 text-2xl font-black text-[#fff7e8]">
                            {game.Round}
                        </div>
                    </div>

                    <div
                        class="rounded-2xl bg-[#f2c36b] p-3 text-center text-[#142833]"
                    >
                        <div
                            class="text-[10px] font-black uppercase tracking-wider opacity-70"
                        >
                            Phase
                        </div>
                        <div class="mt-1 text-sm font-black">
                            {currentPhaseName}
                        </div>
                    </div>

                    <div
                        class={[
                            "rounded-2xl p-3 text-center ring-1",
                            isMyTurn
                                ? "bg-[#73c4bd] text-[#102b38] ring-[#73c4bd]"
                                : "bg-[#f8efe0]/10 text-[#fff7e8] ring-[#f8efe0]/10",
                        ].join(" ")}
                    >
                        <div
                            class="text-[10px] font-black uppercase tracking-wider opacity-70"
                        >
                            Turn
                        </div>
                        <div class="mt-1 text-sm font-black">
                            {currentPlayerName}
                        </div>
                    </div>
                </div>
            </section>

            <section
                class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
            >
                <h2 class="text-xl font-black text-[#fff7e8]">Players</h2>

                <div class="mt-4 space-y-3">
                    {#each game.Players as player}
                        <div
                            class={[
                                "rounded-2xl p-4 ring-1",
                                player.Id === game.CurrentPlayer
                                    ? "bg-[#f2c36b]/16 ring-[#f2c36b]/45"
                                    : "bg-[#f8efe0]/10 ring-[#f8efe0]/10",
                            ].join(" ")}
                        >
                            <div class="flex items-center gap-3">
                                <div
                                    class={[
                                        "grid h-10 w-10 place-items-center rounded-2xl text-sm font-black text-white",
                                        playerColor(player.Id),
                                    ].join(" ")}
                                >
                                    P{player.Id}
                                </div>

                                <div>
                                    <div class="font-black text-[#fff7e8]">
                                        {player.Id === playerId
                                            ? "You"
                                            : `Player ${player.Id}`}
                                    </div>

                                    <div
                                        class="mt-1 text-xs font-semibold text-[#9fc9c5]"
                                    >
                                        {player.Id === game.CurrentPlayer
                                            ? "Taking turn"
                                            : "Waiting"}
                                    </div>
                                </div>
                            </div>

                            <div class="mt-3">
                                <HandCard item={player.Hand} size="sm" />
                            </div>

                            <div
                                class="mt-3 grid grid-cols-3 gap-2 text-center text-xs font-black"
                            >
                                <div
                                    class="rounded-xl bg-[#5b9368]/35 px-2 py-2"
                                    title="Wood"
                                >
                                    W<br />{resourceAmount(player, 1)}
                                </div>

                                <div
                                    class="rounded-xl bg-[#a8adb2]/35 px-2 py-2"
                                    title="Stone"
                                >
                                    S<br />{resourceAmount(player, 2)}
                                </div>

                                <div
                                    class="rounded-xl bg-[#d9b56a]/35 px-2 py-2"
                                    title="Grain"
                                >
                                    G<br />{resourceAmount(player, 3)}
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            </section>
        </aside>

        <Board
            {game}
            {playerId}
            {role}
            {selectedBuildAction}
            {onPlaceTile}
            {onUseDraft}
            onBuild={handleBuild}
        />

        <aside class="space-y-5">
            {#if isMyTurn && game.CurrentPhase !== GamePhase.Pick}
                <section
                    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
                >
                    <div class="flex items-center justify-between gap-3">
                        <h2 class="text-xl font-black text-[#fff7e8]">
                            {game.CurrentPhase === GamePhase.Place
                                ? "Use"
                                : "Build"}
                        </h2>

                        <div
                            class="rounded-xl bg-[#73c4bd] px-3 py-1 text-xs font-black uppercase tracking-wider text-[#102b38]"
                        >
                            Your turn
                        </div>
                    </div>

                    {#if game.CurrentPhase === GamePhase.Place}
                        {#if me?.Hand}
                            <div class="mt-4">
                                <HandCard item={me.Hand} size="lg" />
                            </div>
                        {/if}

                        {#if canPassPlace}
                            <button
                                class="mt-4 w-full cursor-pointer rounded-2xl bg-[#b94b3f] px-5 py-3 font-black text-white shadow-[0_6px_0_rgba(0,0,0,0.18)] transition hover:bg-[#c9574a] active:translate-y-1"
                                type="button"
                                on:click={onPassPlace}
                            >
                                Discard
                            </button>
                        {/if}
                    {/if}

                    {#if game.CurrentPhase === GamePhase.Build}
                        {#if selectedBuildAction}
                            <div
                                class="mt-4 rounded-2xl bg-[#f2c36b]/20 p-3 text-center text-sm font-black text-[#f8efe0] ring-1 ring-[#f2c36b]/40"
                            >
                                {selectedBuildAction === "outpost"
                                    ? "Click a land tile"
                                    : "Click your plain outpost"}
                            </div>
                        {/if}

                        <div class="mt-4 grid grid-cols-2 gap-3">
                            <button
                                class={[
                                    "cursor-pointer rounded-2xl p-4 text-center font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                                    selectedBuildAction === "outpost"
                                        ? "bg-[#f2c36b] text-[#142833]"
                                        : "bg-[#f8efe0]/10 text-[#fff7e8] ring-1 ring-[#f8efe0]/20 hover:bg-[#f8efe0]/16",
                                ].join(" ")}
                                type="button"
                                title="Build Outpost"
                                on:click={() => {
                                    selectedBuildAction = "outpost";
                                    debugLog("build.select", {
                                        action: "outpost",
                                        playerId,
                                        currentPlayer: game.CurrentPlayer,
                                        currentPhase: game.CurrentPhase,
                                        isMyTurn,
                                    });
                                }}
                            >
                                <div class="text-3xl">⌂</div>
                                <div
                                    class="mt-1 text-xs uppercase tracking-wider"
                                >
                                    Outpost
                                </div>
                            </button>

                            <button
                                class={[
                                    "cursor-pointer rounded-2xl p-4 text-center font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                                    selectedBuildAction === "city"
                                        ? "bg-[#f2c36b] text-[#142833]"
                                        : "bg-[#f8efe0]/10 text-[#fff7e8] ring-1 ring-[#f8efe0]/20 hover:bg-[#f8efe0]/16",
                                ].join(" ")}
                                type="button"
                                title="Upgrade City"
                                on:click={() => {
                                    selectedBuildAction = "city";
                                    debugLog("build.select", {
                                        action: "city",
                                        playerId,
                                        currentPlayer: game.CurrentPlayer,
                                        currentPhase: game.CurrentPhase,
                                        isMyTurn,
                                    });
                                }}
                            >
                                <div class="text-3xl">▦</div>
                                <div
                                    class="mt-1 text-xs uppercase tracking-wider"
                                >
                                    City
                                </div>
                            </button>
                        </div>

                        <button
                            class="mt-4 w-full cursor-pointer rounded-2xl bg-[#f8efe0]/10 px-5 py-3 font-black text-[#fff7e8] shadow-[0_6px_0_rgba(0,0,0,0.18)] ring-1 ring-[#f8efe0]/20 transition hover:bg-[#f8efe0]/16 active:translate-y-1"
                            type="button"
                            on:click={onPassBuild}
                        >
                            Pass
                        </button>
                    {/if}

                    {#if error}
                        <div
                            class="mt-4 rounded-2xl bg-[#b94b3f] px-5 py-3 text-sm font-semibold text-white"
                        >
                            {error}
                        </div>
                    {/if}
                </section>
            {:else if error}
                <section
                    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
                >
                    <div
                        class="rounded-2xl bg-[#b94b3f] px-5 py-3 text-sm font-semibold text-white"
                    >
                        {error}
                    </div>
                </section>
            {/if}

            <Market {game} {playerId} {role} {onPick} />
        </aside>
    </section>
</main>

<style>
    .logo-diamond {
        display: block;
        line-height: 1;
        transform: translateY(-1px);
    }

    .bg-board-flat {
        background: linear-gradient(180deg, #173943 0%, #102832 100%);
    }

    .bg-board-texture {
        background-image: radial-gradient(
            circle,
            rgba(255, 255, 255, 0.055) 1px,
            transparent 1px
        );
        background-size: 28px 28px;
    }
</style>
