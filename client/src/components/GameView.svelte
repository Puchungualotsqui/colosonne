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

    function phaseName(phase: GamePhase) {
        switch (phase) {
            case GamePhase.Pick:
                return "Draft";
            case GamePhase.Place:
                return "Place / Use";
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

    function instruction() {
        if (role === "spectator") {
            return "You are watching this match. Spectators cannot make moves.";
        }

        if (!isMyTurn) {
            return `Waiting for ${currentPlayerName}.`;
        }

        if (game.CurrentPhase === GamePhase.Pick) {
            return "Choose one card from the market.";
        }

        if (game.CurrentPhase === GamePhase.Place) {
            if (!me?.Hand) {
                return "You have no draft item. Pass the place/use step.";
            }

            if (me.Hand.Kind === DraftKind.Tile) {
                return "Click a highlighted empty hex to place your tile.";
            }

            return "Click a board tile to use your drafted item, or pass if it is not usable.";
        }

        if (game.CurrentPhase === GamePhase.Build) {
            if (selectedBuildAction === "outpost") {
                return "Click a land tile to build an outpost.";
            }

            if (selectedBuildAction === "city") {
                return "Click one of your plain outposts to upgrade it into a city.";
            }

            return "Choose a build action, or pass.";
        }

        return "Waiting for game state.";
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

                <div
                    class="mt-4 space-y-3 text-sm font-semibold text-[#d9e6df]"
                >
                    <div class="flex justify-between gap-3">
                        <span>Round</span>
                        <span class="font-black text-[#fff7e8]"
                            >{game.Round}</span
                        >
                    </div>

                    <div class="flex justify-between gap-3">
                        <span>Phase</span>
                        <span class="font-black text-[#fff7e8]"
                            >{currentPhaseName}</span
                        >
                    </div>

                    <div class="flex justify-between gap-3">
                        <span>Turn</span>
                        <span class="font-black text-[#fff7e8]"
                            >{currentPlayerName}</span
                        >
                    </div>

                    <div class="flex justify-between gap-3">
                        <span>Your role</span>
                        <span class="font-black capitalize text-[#fff7e8]"
                            >{role || "unknown"}</span
                        >
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
                            class="rounded-2xl bg-[#f8efe0]/10 p-4 ring-1 ring-[#f8efe0]/10"
                        >
                            <div
                                class="flex items-center justify-between gap-3"
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
                                            class="text-xs font-semibold text-[#9fc9c5]"
                                        >
                                            Hand: {draftName(player.Hand)}
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div
                                class="mt-3 grid grid-cols-3 gap-2 text-center text-xs font-black"
                            >
                                <div
                                    class="rounded-xl bg-[#5b9368]/35 px-2 py-2"
                                >
                                    Wood<br />{resourceAmount(player, 1)}
                                </div>
                                <div
                                    class="rounded-xl bg-[#a8adb2]/35 px-2 py-2"
                                >
                                    Stone<br />{resourceAmount(player, 2)}
                                </div>
                                <div
                                    class="rounded-xl bg-[#d9b56a]/35 px-2 py-2"
                                >
                                    Grain<br />{resourceAmount(player, 3)}
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
            <section
                class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
            >
                <h2 class="text-xl font-black text-[#fff7e8]">Action</h2>

                <div
                    class="mt-3 rounded-2xl bg-[#f8efe0]/10 p-4 text-sm font-semibold leading-6 text-[#d9e6df] ring-1 ring-[#f8efe0]/10"
                >
                    {instruction()}
                </div>

                {#if game.CurrentPhase === GamePhase.Place && isMyTurn}
                    <button
                        class="mt-4 w-full cursor-pointer rounded-2xl bg-[#f8efe0]/10 px-5 py-3 font-black text-[#fff7e8] shadow-[0_6px_0_rgba(0,0,0,0.18)] ring-1 ring-[#f8efe0]/20 transition hover:bg-[#f8efe0]/16 active:translate-y-1"
                        type="button"
                        on:click={onPassPlace}
                    >
                        Pass Place / Use
                    </button>
                {/if}

                {#if game.CurrentPhase === GamePhase.Build && isMyTurn}
                    <div class="mt-4 grid gap-3">
                        <button
                            class={[
                                "cursor-pointer rounded-2xl px-5 py-3 font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                                selectedBuildAction === "outpost"
                                    ? "bg-[#f2c36b] text-[#142833]"
                                    : "bg-[#f8efe0]/10 text-[#fff7e8] ring-1 ring-[#f8efe0]/20 hover:bg-[#f8efe0]/16",
                            ].join(" ")}
                            type="button"
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
                            Build Outpost
                        </button>

                        <button
                            class={[
                                "cursor-pointer rounded-2xl px-5 py-3 font-black shadow-[0_6px_0_rgba(0,0,0,0.18)] transition active:translate-y-1",
                                selectedBuildAction === "city"
                                    ? "bg-[#f2c36b] text-[#142833]"
                                    : "bg-[#f8efe0]/10 text-[#fff7e8] ring-1 ring-[#f8efe0]/20 hover:bg-[#f8efe0]/16",
                            ].join(" ")}
                            type="button"
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
                            Upgrade City
                        </button>

                        <button
                            class="cursor-pointer rounded-2xl bg-[#b94b3f] px-5 py-3 font-black text-white shadow-[0_6px_0_rgba(0,0,0,0.18)] transition hover:bg-[#c9574a] active:translate-y-1"
                            type="button"
                            on:click={onPassBuild}
                        >
                            Pass Build
                        </button>
                    </div>
                {/if}

                {#if error}
                    <div
                        class="mt-4 rounded-2xl bg-[#b94b3f] px-5 py-3 font-semibold text-white"
                    >
                        {error}
                    </div>
                {/if}
            </section>

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
