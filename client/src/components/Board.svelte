<script lang="ts">
    import {
        Biome,
        DraftKind,
        GamePhase,
        Structure,
        type GameState,
        type Tile,
    } from "../lib/types";
    import { debugLog } from "../lib/debug";

    export let game: GameState;
    export let playerId = 0;
    export let role: "player" | "spectator" | "" = "";
    export let selectedBuildAction: "outpost" | "city" | null = null;

    export let onPlaceTile: (x: number, y: number) => void;
    export let onUseDraft: (x: number, y: number) => void;
    export let onBuild: (
        action: "outpost" | "city",
        x: number,
        y: number,
    ) => void;

    type Coord = {
        x: number;
        y: number;
    };

    type RenderHex = Coord & {
        key: string;
        left: number;
        top: number;
        tile?: Tile;
        candidate: boolean;
    };

    const HEX_W = 112;
    const HEX_H = 98;
    const STEP_X = 93;
    const STEP_Y = 74;
    const PAD = 90;

    $: me = game.Players.find((p) => p.Id === playerId);
    $: hand = me?.Hand ?? null;

    $: isMyTurn = role === "player" && game.CurrentPlayer === playerId;
    $: canPlaceTile =
        isMyTurn &&
        game.CurrentPhase === GamePhase.Place &&
        hand?.Kind === DraftKind.Tile;

    $: canUseDraft =
        isMyTurn &&
        game.CurrentPhase === GamePhase.Place &&
        hand !== null &&
        hand.Kind !== DraftKind.Tile;

    $: canBuild =
        isMyTurn &&
        game.CurrentPhase === GamePhase.Build &&
        selectedBuildAction !== null;

    $: candidates = canPlaceTile ? placementCandidates(game.Map) : [];
    $: renderHexes = buildRenderHexes(game.Map, candidates);
    $: boardSize = calculateBoardSize(renderHexes);

    $: debugLog("board.state", {
        role,
        playerId,
        currentPlayer: game.CurrentPlayer,
        currentPhase: game.CurrentPhase,
        isMyTurn,
        hand,
        canPlaceTile,
        canUseDraft,
        canBuild,
        selectedBuildAction,
        candidatesCount: candidates.length,
    });

    function key(x: number, y: number) {
        return `${x},${y}`;
    }

    function hexNeighbors(x: number, y: number): Coord[] {
        return [
            { x: x + 1, y },
            { x: x + 1, y: y - 1 },
            { x, y: y - 1 },
            { x: x - 1, y },
            { x: x - 1, y: y + 1 },
            { x, y: y + 1 },
        ];
    }

    function rawPosition(x: number, y: number) {
        return {
            left: (x + y / 2) * STEP_X,
            top: y * STEP_Y,
        };
    }

    function placementCandidates(map: Tile[]): Coord[] {
        const occupied = new Set(map.map((t) => key(t.X, t.Y)));
        const out = new Map<string, Coord>();

        for (const tile of map) {
            for (const n of hexNeighbors(tile.X, tile.Y)) {
                const k = key(n.x, n.y);
                if (!occupied.has(k)) {
                    out.set(k, n);
                }
            }
        }

        return [...out.values()];
    }

    function buildRenderHexes(
        map: Tile[],
        candidateCoords: Coord[],
    ): RenderHex[] {
        const tileByKey = new Map(map.map((t) => [key(t.X, t.Y), t]));
        const all = [
            ...map.map((tile) => ({
                x: tile.X,
                y: tile.Y,
                tile,
                candidate: false,
            })),
            ...candidateCoords.map((c) => ({
                x: c.x,
                y: c.y,
                tile: undefined,
                candidate: true,
            })),
        ];

        const raw = all.map((item) => {
            const pos = rawPosition(item.x, item.y);
            return {
                ...item,
                key: key(item.x, item.y),
                left: pos.left,
                top: pos.top,
            };
        });

        const minLeft = Math.min(...raw.map((h) => h.left));
        const minTop = Math.min(...raw.map((h) => h.top));

        return raw.map((h) => ({
            ...h,
            left: h.left - minLeft + PAD,
            top: h.top - minTop + PAD,
            tile: tileByKey.get(h.key),
        }));
    }

    function calculateBoardSize(hexes: RenderHex[]) {
        if (hexes.length === 0) {
            return {
                width: 600,
                height: 440,
            };
        }

        const maxLeft = Math.max(...hexes.map((h) => h.left));
        const maxTop = Math.max(...hexes.map((h) => h.top));

        return {
            width: Math.max(620, maxLeft + HEX_W + PAD),
            height: Math.max(440, maxTop + HEX_H + PAD),
        };
    }

    function biomeClass(tile: Tile | undefined, candidate: boolean) {
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
            default:
                return "border-[#6b4a2f] bg-[#ead7aa] text-[#142833]";
        }
    }

    function biomeLabel(biome: Biome | undefined) {
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
                return "";
        }
    }

    function structureLabel(structure: Structure) {
        switch (structure) {
            case Structure.Outpost:
                return "Outpost";
            case Structure.City:
                return "City";
            case Structure.Bridge:
                return "Bridge";
            case Structure.Watchtower:
                return "Watchtower";
            case Structure.Road:
                return "Road";
            default:
                return "";
        }
    }

    function structureIcon(structure: Structure) {
        switch (structure) {
            case Structure.Outpost:
                return "⌂";
            case Structure.City:
                return "▦";
            case Structure.Bridge:
                return "⌒";
            case Structure.Watchtower:
                return "♜";
            case Structure.Road:
                return "━";
            default:
                return "";
        }
    }

    function ownerClass(tile: Tile) {
        if (!tile.HasOwner) return "bg-[#f8efe0]/80 text-[#5c4934]";

        if (tile.Owner === 1) {
            return "bg-[#1d4e89] text-white";
        }

        if (tile.Owner === 2) {
            return "bg-[#b94b3f] text-white";
        }

        return "bg-[#f8efe0]/80 text-[#5c4934]";
    }

    function ownerLabel(tile: Tile) {
        if (!tile.HasOwner) return "Open";
        return `P${tile.Owner}`;
    }

    function tileTooltip(tile: Tile | undefined, candidate: boolean) {
        if (candidate) {
            return "Empty frontier hex";
        }

        if (!tile) {
            return "";
        }

        const owner = tile.HasOwner ? `P${tile.Owner}` : "Open";
        const structure =
            tile.Structure !== Structure.None
                ? structureLabel(tile.Structure)
                : "No structure";

        return `${biomeLabel(tile.Biome)} · ${owner} · ${structure}`;
    }

    function handleHexClick(hex: RenderHex) {
        debugLog("board.hex.click", {
            x: hex.x,
            y: hex.y,
            hasTile: !!hex.tile,
            candidate: hex.candidate,
            role,
            playerId,
            currentPlayer: game.CurrentPlayer,
            currentPhase: game.CurrentPhase,
            isMyTurn,
            hand,
            canPlaceTile,
            canUseDraft,
            canBuild,
            selectedBuildAction,
        });

        if (hex.candidate && canPlaceTile) {
            debugLog("board.place_tile.send", {
                x: hex.x,
                y: hex.y,
            });

            onPlaceTile(hex.x, hex.y);
            return;
        }

        if (!hex.tile) return;

        if (canUseDraft) {
            debugLog("board.use_draft.send", {
                x: hex.tile.X,
                y: hex.tile.Y,
            });

            onUseDraft(hex.tile.X, hex.tile.Y);
            return;
        }

        if (canBuild && selectedBuildAction) {
            debugLog("board.build.send", {
                action: selectedBuildAction,
                x: hex.tile.X,
                y: hex.tile.Y,
            });

            onBuild(selectedBuildAction, hex.tile.X, hex.tile.Y);
        }
    }

    function isClickable(hex: RenderHex) {
        if (hex.candidate && canPlaceTile) return true;
        if (hex.tile && canUseDraft) return true;
        if (hex.tile && canBuild) return true;
        return false;
    }
</script>

<section
    class="rounded-[34px] bg-[#caa66d] p-4 shadow-[0_18px_0_rgba(44,31,21,0.28)] ring-1 ring-black/20"
>
    <div
        class="rounded-[26px] border border-[#6b4a2f]/35 bg-[#ead7aa] p-4 shadow-inner"
    >
        <div class="mb-4 flex items-center justify-between gap-3">
            <h2 class="text-xl font-black text-[#142833]">Board</h2>

            <div
                class="rounded-xl bg-[#f8efe0] px-3 py-1 text-sm font-black text-[#142833]"
            >
                {game.Map.length} tiles
            </div>
        </div>

        <div class="overflow-auto rounded-2xl bg-[#d9c291] p-4 shadow-inner">
            <div
                class="relative"
                style={`width: ${boardSize.width}px; height: ${boardSize.height}px;`}
            >
                {#each renderHexes as hex}
                    <button
                        class={[
                            "group clip-hex absolute flex items-center justify-center border-[2px] shadow-[0_7px_0_rgba(74,48,31,0.22)] transition",
                            biomeClass(hex.tile, hex.candidate),
                            isClickable(hex)
                                ? "cursor-pointer hover:-translate-y-1 hover:brightness-110"
                                : "cursor-default",
                            hex.candidate ? "border-dashed opacity-80" : "",
                        ].join(" ")}
                        style={`left: ${hex.left}px; top: ${hex.top}px; width: ${HEX_W}px; height: ${HEX_H}px;`}
                        type="button"
                        disabled={!isClickable(hex)}
                        aria-label={tileTooltip(hex.tile, hex.candidate)}
                        on:click={() => handleHexClick(hex)}
                    >
                        <div
                            class="pointer-events-none absolute inset-[5px] clip-hex border border-white/25"
                        ></div>

                        {#if hex.candidate}
                            <div
                                class="relative z-10 grid h-10 w-10 place-items-center rounded-full bg-[#f8efe0]/70 text-2xl font-black text-[#6b4a2f]"
                            >
                                +
                            </div>
                        {:else if hex.tile}
                            {#if hex.tile.Structure !== Structure.None}
                                <div
                                    class="relative z-10 grid h-12 w-12 place-items-center rounded-2xl bg-[#f8efe0]/70 text-2xl font-black text-[#142833] shadow-sm"
                                    title={structureLabel(hex.tile.Structure)}
                                >
                                    {structureIcon(hex.tile.Structure)}
                                </div>
                            {/if}

                            {#if hex.tile.HasOwner}
                                <div
                                    class={[
                                        "absolute bottom-2 left-1/2 z-20 h-5 min-w-8 -translate-x-1/2 rounded-full px-2 text-[10px] font-black leading-5 shadow-sm",
                                        ownerClass(hex.tile),
                                    ].join(" ")}
                                >
                                    {ownerLabel(hex.tile)}
                                </div>
                            {/if}
                        {/if}

                        <div
                            class="pointer-events-none absolute bottom-[calc(100%+8px)] left-1/2 z-50 hidden w-max max-w-[220px] -translate-x-1/2 rounded-xl bg-[#142833] px-3 py-2 text-xs font-bold text-[#f8efe0] shadow-xl ring-1 ring-white/10 group-hover:block"
                        >
                            {tileTooltip(hex.tile, hex.candidate)}
                        </div>
                    </button>
                {/each}
            </div>
        </div>
    </div>
</section>

<style>
    .clip-hex {
        clip-path: polygon(
            50% 0%,
            93.3% 25%,
            93.3% 75%,
            50% 100%,
            6.7% 75%,
            6.7% 25%
        );
    }
</style>
