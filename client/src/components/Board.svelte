<script lang="ts">
    import {
        Action,
        Biome,
        DraftKind,
        GamePhase,
        Structure,
        type DraftItem,
        type GameState,
        type TargetBuildAction,
        type Tile,
    } from "../lib/types";
    import { debugLog } from "../lib/debug";
    import BoardTile from "./BoardTile.svelte";

    export let game: GameState;
    export let playerId = 0;
    export let role: "player" | "spectator" | "" = "";
    export let selectedBuildAction: TargetBuildAction | null = null;

    export let selectedHandIndex = -1;
    export let selectedHandItem: DraftItem | null = null;

    export let onPlaceTile: (handIndex: number, x: number, y: number) => void;
    export let onUseDraft: (
        handIndex: number,
        x: number,
        y: number,
        targetPlayerId?: number,
    ) => void;
    export let onBuild: (
        action: TargetBuildAction,
        x: number,
        y: number,
    ) => void;

    type InfluenceTooltipRow = {
        playerId: number;
        influence: number;
        tempInfluence: number;
        total: number;
        leading: boolean;
    };

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
        ghost: boolean;
    };

    type InfluenceLeader = {
        playerId: number;
        value: number;
        tied: boolean;
        total: number;
    };

    type TargetIntent =
        | { kind: "none" }
        | { kind: "placeTile"; handIndex: number }
        | { kind: "draft"; handIndex: number; item: DraftItem }
        | { kind: "build"; action: TargetBuildAction };

    type HexTargetStatus = {
        clickable: boolean;
        dimmed: boolean;
    };

    const HEX_W = 112;
    const HEX_H = 98;
    const STEP_X = 93;
    const STEP_Y = 74;
    const PAD = 90;

    $: isMyTurn = role === "player" && game.CurrentPlayer === playerId;

    $: canPlaceTile =
        isMyTurn &&
        game.CurrentPhase === GamePhase.Place &&
        selectedHandIndex >= 0 &&
        selectedHandItem?.Kind === DraftKind.Tile;

    $: canUseDraft =
        isMyTurn &&
        game.CurrentPhase === GamePhase.Place &&
        selectedHandIndex >= 0 &&
        selectedHandItem !== null &&
        selectedHandItem.Kind !== DraftKind.Tile &&
        draftNeedsBoardTarget(selectedHandItem);

    $: canBuild =
        isMyTurn &&
        game.CurrentPhase === GamePhase.Build &&
        selectedBuildAction !== null;

    $: placementShell = placementCandidates(game.Map);
    $: renderHexes = buildRenderHexes(game.Map, placementShell, canPlaceTile);
    $: boardSize = calculateBoardSize(renderHexes);

    let targetIntent: TargetIntent = { kind: "none" };

    $: targetIntent = getTargetIntent();

    $: debugLog("board.targeting.state", {
        isMyTurn,
        currentPhase: game.CurrentPhase,
        selectedBuildAction,
        selectedHandIndex,
        selectedHandItem,
        targetIntent,
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
        showCandidates: boolean,
    ): RenderHex[] {
        const tileByKey = new Map(map.map((t) => [key(t.X, t.Y), t]));

        const all = [
            ...map.map((tile) => ({
                x: tile.X,
                y: tile.Y,
                tile,
                candidate: false,
                ghost: false,
            })),
            ...candidateCoords.map((c) => ({
                x: c.x,
                y: c.y,
                tile: undefined,
                candidate: showCandidates,
                ghost: !showCandidates,
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

    function influenceEntries(tile: Tile | undefined) {
        if (!tile) return [];

        return Object.entries(tile.Influence ?? {})
            .map(([playerId, value]) => ({
                playerId: Number(playerId),
                value: Number(value),
            }))
            .filter((entry) => entry.value > 0);
    }

    function influenceLeader(tile: Tile | undefined): InfluenceLeader | null {
        const entries = influenceEntries(tile);

        if (entries.length === 0) return null;

        let bestPlayer = 0;
        let bestValue = 0;
        let tied = false;
        let total = 0;

        for (const entry of entries) {
            total += entry.value;

            if (entry.value > bestValue) {
                bestPlayer = entry.playerId;
                bestValue = entry.value;
                tied = false;
            } else if (entry.value === bestValue) {
                tied = true;
            }
        }

        return {
            playerId: bestPlayer,
            value: bestValue,
            tied,
            total,
        };
    }

    function tileAuraOwner(tile: Tile | undefined) {
        if (!tile) return 0;

        const leader = influenceLeader(tile);

        // If visible influence exists, use it.
        if (leader) {
            if (leader.tied) return 3;
            if (leader.playerId === 1) return 1;
            if (leader.playerId === 2) return 2;
        }

        // Fallback: use ownership.
        // This is what makes the empire outline visible even after Influence is empty
        // or only recalculated temporarily.
        if (tile.HasOwner) {
            if (tile.Owner === 1) return 1;
            if (tile.Owner === 2) return 2;
        }

        return 0;
    }

    function sameAuraGroup(a: Tile | undefined, b: Tile | undefined) {
        const ownerA = tileAuraOwner(a);
        const ownerB = tileAuraOwner(b);

        // Tied influence should not connect into a player empire mass.
        if (ownerA === 3 || ownerB === 3) return false;

        return ownerA > 0 && ownerA === ownerB;
    }

    function auraEdgesForTile(tile: Tile | undefined) {
        if (!tile) return [false, false, false, false, false, false];

        const owner = tileAuraOwner(tile);
        if (owner === 0) return [false, false, false, false, false, false];

        // Neighbor order from hexNeighbors:
        // 0 E, 1 NE, 2 NW, 3 W, 4 SW, 5 SE
        //
        // Edge order in BoardTile:
        // 0 top-right, 1 right, 2 bottom-right, 3 bottom-left, 4 left, 5 top-left
        const neighborToEdge = [1, 0, 5, 4, 3, 2];

        const edges = [false, false, false, false, false, false];
        const neighbors = hexNeighbors(tile.X, tile.Y);

        for (let i = 0; i < neighbors.length; i++) {
            const n = neighbors[i];
            const neighborTile = game.Map.find(
                (t) => t.X === n.x && t.Y === n.y,
            );
            const edgeIndex = neighborToEdge[i];

            // Show an edge only when this side touches non-matching territory,
            // no tile, enemy tile, neutral tile, or tied area.
            edges[edgeIndex] = !sameAuraGroup(tile, neighborTile);
        }

        return edges;
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
            case Biome.Ruins:
                return "border-[#6d4c9b] bg-[#9b79c9] text-[#142833]";
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
            case Biome.Ruins:
                return "Ruins";
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

    function structureIcon(structure: Structure) {
        switch (structure) {
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

    function controlsTile(
        tile: { HasOwner: boolean; Owner: number } | undefined,
    ) {
        return !!tile && tile.HasOwner && tile.Owner === playerId;
    }

    function hasAdjacentControlledTile(x: number, y: number) {
        return hexNeighbors(x, y).some((n) => {
            const tile = game.Map.find((t) => t.X === n.x && t.Y === n.y);
            return controlsTile(tile);
        });
    }

    function canBuildBridgeOnTile(tile: Tile | undefined) {
        if (!tile) return false;

        return tile.Biome === Biome.River && tile.Structure === Structure.None;
    }

    function isEnemyControlledTile(tile: Tile | undefined) {
        return !!tile && tile.HasOwner && tile.Owner !== playerId;
    }

    function isUnownedTile(tile: Tile | undefined) {
        return !!tile && !tile.HasOwner;
    }

    function canBuildBlockadeOnTile(tile: Tile | undefined) {
        if (!tile) return false;

        // Blockade cannot be placed on rivers.
        if (tile.Biome === Biome.River) return false;

        // Cannot place on already blockaded tile.
        if (tile.HasBlockade) return false;

        // Cannot blockade your own tile.
        if (controlsTile(tile)) return false;

        // Enemy tile is valid, with or without a structure.
        if (isEnemyControlledTile(tile)) return true;

        // Neutral tile is valid only if adjacent to your controlled territory.
        if (isUnownedTile(tile)) {
            return hasAdjacentControlledTile(tile.X, tile.Y);
        }

        return false;
    }

    function blockadeTargetMessage(tile: Tile) {
        if (tile.Biome === Biome.River) return "Cannot blockade river";
        if (tile.HasBlockade) return "Already blockaded";
        if (controlsTile(tile)) return "Cannot blockade your own tile";

        if (isEnemyControlledTile(tile)) {
            return tile.Structure !== Structure.None
                ? "Blockade enemy structure"
                : "Blockade enemy tile";
        }

        if (isUnownedTile(tile)) {
            return hasAdjacentControlledTile(tile.X, tile.Y)
                ? "Blockade neutral tile"
                : "Neutral blockade needs adjacent controlled tile";
        }

        return "Invalid blockade target";
    }

    function settlementTargetMessage(tile: Tile) {
        if (tile.Biome === Biome.River)
            return "Settlement cannot be built on river";
        if (tile.Structure !== Structure.None)
            return "Tile already has a structure";
        if (!controlsTile(tile))
            return "Settlement requires friendly territory";
        return "Build Settlement";
    }

    function draftNeedsBoardTarget(item: DraftItem) {
        if (item.Kind === DraftKind.Structure) return true;

        if (item.Kind === DraftKind.Action) {
            return (
                item.Action === Action.Harvest ||
                item.Action === Action.Reinforce
            );
        }

        return false;
    }

    function isValidDraftTarget(
        item: DraftItem | null,
        tile: Tile | undefined,
    ) {
        if (!item || !tile) return false;

        if (item.Kind === DraftKind.Structure) {
            if (tile.Structure !== Structure.None) return false;

            switch (item.Structure) {
                case Structure.Bridge:
                    return canBuildBridgeOnTile(tile);

                case Structure.Watchtower:
                    return (
                        tile.Biome !== Biome.River &&
                        tile.Structure === Structure.None &&
                        !isEnemyControlledTile(tile)
                    );

                case Structure.Outpost:
                case Structure.City:
                case Structure.Settlement:
                    return false;

                default:
                    return false;
            }
        }

        if (item.Kind === DraftKind.Action) {
            if (item.Action === Action.Harvest) {
                return controlsTile(tile) && tile.Biome !== Biome.River;
            }

            if (item.Action === Action.Reinforce) {
                return true;
            }
        }

        return false;
    }

    function canBuildOutpostOnTile(tile: Tile | undefined) {
        if (!tile) return false;

        return tile.Biome !== Biome.River && tile.Structure === Structure.None;
    }

    function canBuildSettlementOnTile(tile: Tile | undefined) {
        if (!tile) return false;

        return (
            tile.Biome !== Biome.River &&
            tile.Structure === Structure.None &&
            controlsTile(tile)
        );
    }

    function canUpgradeCityOnTile(tile: Tile | undefined) {
        if (!tile) return false;

        return (
            tile.Structure === Structure.Outpost &&
            tile.StructureOwner === playerId
        );
    }

    function canUseFloodOnTile(tile: Tile | undefined) {
        if (!tile) return false;

        return tile.Biome !== Biome.River && tile.Structure === Structure.None;
    }

    function isValidBuildTarget(tile: Tile | undefined) {
        if (!canBuild || !selectedBuildAction || !tile) return false;

        switch (selectedBuildAction) {
            case "outpost":
                return canBuildOutpostOnTile(tile);

            case "settlement":
                return canBuildSettlementOnTile(tile);

            case "city":
                return canUpgradeCityOnTile(tile);

            case "blockade":
                return canBuildBlockadeOnTile(tile);

            case "flood":
                return canUseFloodOnTile(tile);

            default:
                return false;
        }
    }

    function tileTooltip(tile: Tile | undefined, candidate: boolean) {
        if (candidate) return "Place tile";
        if (!tile) return "";

        if (canUseDraft && selectedHandItem) {
            if (isValidDraftTarget(selectedHandItem, tile)) {
                if (
                    selectedHandItem.Kind === DraftKind.Structure &&
                    selectedHandItem.Structure === Structure.Watchtower
                ) {
                    return tile.HasOwner
                        ? "Build Watchtower on your tile"
                        : "Build Watchtower on neutral tile";
                }

                if (
                    selectedHandItem.Kind === DraftKind.Structure &&
                    selectedHandItem.Structure === Structure.Bridge
                ) {
                    return "Build Bridge";
                }

                return "Use card here";
            }

            if (
                selectedHandItem.Kind === DraftKind.Structure &&
                selectedHandItem.Structure === Structure.Watchtower
            ) {
                if (tile.Biome === Biome.River)
                    return "Watchtower cannot be built on river";
                if (tile.Structure !== Structure.None)
                    return "Tile already has a structure";
                if (isEnemyControlledTile(tile))
                    return "Cannot build Watchtower on enemy tile";
            }
        }

        if (canBuild && selectedBuildAction) {
            if (isValidBuildTarget(tile)) {
                switch (selectedBuildAction) {
                    case "outpost":
                        return "Build Outpost";

                    case "settlement":
                        return "Build Settlement";

                    case "city":
                        return "Upgrade Outpost to City";

                    case "blockade":
                        return blockadeTargetMessage(tile);

                    case "flood":
                        return "Convert to River";
                }
            }

            switch (selectedBuildAction) {
                case "outpost":
                    if (tile.Biome === Biome.River) {
                        return "Outpost cannot be built on River";
                    }

                    if (tile.Structure !== Structure.None) {
                        return "Tile already has a structure";
                    }

                    return "Invalid Outpost target";

                case "settlement":
                    return settlementTargetMessage(tile);

                case "city":
                    if (tile.Structure === Structure.Outpost) {
                        return tile.StructureOwner === playerId
                            ? "Upgrade Outpost to City"
                            : "Requires your own Outpost";
                    }

                    return "City requires your own Outpost";

                case "blockade":
                    return blockadeTargetMessage(tile);

                case "flood":
                    if (tile.Biome === Biome.River) {
                        return "Already river";
                    }

                    if (tile.Structure !== Structure.None) {
                        return "Cannot flood a structure";
                    }

                    return "Invalid Flood target";
            }
        }

        const owner = tile.HasOwner ? `P${tile.Owner}` : "Open";
        const structure =
            tile.Structure !== Structure.None
                ? structureLabel(tile.Structure)
                : "No structure";

        if (tile.Biome === Biome.River && tile.Structure !== Structure.Bridge) {
            return `River · Not controllable · ${structure}`;
        }

        if (tile.Biome === Biome.Ruins) {
            return `Ruins · ${owner} · ${structure} · Relic`;
        }

        return `${biomeLabel(tile.Biome)} · ${owner} · ${structure}`;
    }

    function handleHexSelect(hex: RenderHex) {
        const status = getHexTargetStatus(hex);

        debugLog("board.hex.select", {
            x: hex.x,
            y: hex.y,
            hasTile: !!hex.tile,
            candidate: hex.candidate,
            role,
            playerId,
            currentPlayer: game.CurrentPlayer,
            currentPhase: game.CurrentPhase,
            isMyTurn,
            selectedBuildAction,
            selectedHandIndex,
            selectedHandItem,
            targetIntent,
            clickable: status.clickable,
            dimmed: status.dimmed,
            tile: hex.tile,
        });

        if (!status.clickable) return;

        switch (targetIntent.kind) {
            case "placeTile":
                if (!hex.candidate) return;
                onPlaceTile(targetIntent.handIndex, hex.x, hex.y);
                return;

            case "draft":
                if (!hex.tile) return;
                onUseDraft(targetIntent.handIndex, hex.tile.X, hex.tile.Y);
                return;

            case "build":
                if (!hex.tile) return;
                onBuild(targetIntent.action, hex.tile.X, hex.tile.Y);
                return;

            case "none":
                return;
        }
    }

    function getTargetIntent(): TargetIntent {
        if (
            isMyTurn &&
            game.CurrentPhase === GamePhase.Place &&
            selectedHandIndex >= 0 &&
            selectedHandItem?.Kind === DraftKind.Tile
        ) {
            return {
                kind: "placeTile",
                handIndex: selectedHandIndex,
            };
        }

        if (
            isMyTurn &&
            game.CurrentPhase === GamePhase.Place &&
            selectedHandIndex >= 0 &&
            selectedHandItem !== null &&
            selectedHandItem.Kind !== DraftKind.Tile &&
            draftNeedsBoardTarget(selectedHandItem)
        ) {
            return {
                kind: "draft",
                handIndex: selectedHandIndex,
                item: selectedHandItem,
            };
        }

        if (
            isMyTurn &&
            game.CurrentPhase === GamePhase.Build &&
            selectedBuildAction !== null
        ) {
            return {
                kind: "build",
                action: selectedBuildAction,
            };
        }

        return { kind: "none" };
    }

    function isTargetingExistingTile(intent: TargetIntent) {
        return intent.kind === "draft" || intent.kind === "build";
    }

    function isValidTargetForIntent(
        intent: TargetIntent,
        tile: Tile | undefined,
    ) {
        if (!tile) return false;

        switch (intent.kind) {
            case "draft":
                return isValidDraftTarget(intent.item, tile);

            case "build":
                switch (intent.action) {
                    case "outpost":
                        return canBuildOutpostOnTile(tile);

                    case "settlement":
                        return canBuildSettlementOnTile(tile);

                    case "city":
                        return canUpgradeCityOnTile(tile);

                    case "blockade":
                        return canBuildBlockadeOnTile(tile);

                    case "flood":
                        return canUseFloodOnTile(tile);

                    default:
                        return false;
                }

            default:
                return false;
        }
    }

    function getHexTargetStatus(hex: RenderHex): HexTargetStatus {
        const intent = targetIntent;

        if (intent.kind === "none") {
            return {
                clickable: false,
                dimmed: false,
            };
        }

        if (intent.kind === "placeTile") {
            return {
                clickable: hex.candidate,
                dimmed: false,
            };
        }

        if (!hex.tile) {
            return {
                clickable: false,
                dimmed: false,
            };
        }

        const valid = isValidTargetForIntent(intent, hex.tile);

        return {
            clickable: valid,
            dimmed: isTargetingExistingTile(intent) && !valid,
        };
    }

    let hoveredStructureKey = "";
    let hoveredStructureInfluence = new Set<string>();
    let hoveredStructureOwner = 0;

    function tileAt(x: number, y: number) {
        return game.Map.find((tile) => tile.X === x && tile.Y === y);
    }

    function canReceiveInfluencePreview(tile: Tile | undefined) {
        if (!tile) return false;
        if (tile.Biome === Biome.River && tile.Structure !== Structure.Bridge) {
            return false;
        }

        return true;
    }

    function hexRingRadius2(x: number, y: number): Coord[] {
        const results: Coord[] = [];

        for (let dx = -2; dx <= 2; dx++) {
            for (let dy = -2; dy <= 2; dy++) {
                const dz = -dx - dy;

                if (
                    (Math.abs(dx) === 2 ||
                        Math.abs(dy) === 2 ||
                        Math.abs(dz) === 2) &&
                    Math.abs(dx) <= 2 &&
                    Math.abs(dy) <= 2 &&
                    Math.abs(dz) <= 2
                ) {
                    results.push({ x: x + dx, y: y + dy });
                }
            }
        }

        return results;
    }

    function isUnbridgedRiver(tile: Tile | undefined) {
        return (
            !!tile &&
            tile.Biome === Biome.River &&
            tile.Structure !== Structure.Bridge
        );
    }

    function canInfluenceDistance2Preview(
        fromX: number,
        fromY: number,
        toX: number,
        toY: number,
    ) {
        for (const a of hexNeighbors(fromX, fromY)) {
            for (const b of hexNeighbors(toX, toY)) {
                if (a.x !== b.x || a.y !== b.y) continue;

                const middle = tileAt(a.x, a.y);
                if (!middle) continue;

                if (!isUnbridgedRiver(middle)) {
                    return true;
                }
            }
        }

        return false;
    }

    function structureInfluenceCoords(tile: Tile | undefined) {
        if (!tile) return [];

        const coords: Coord[] = [];

        function addCoord(x: number, y: number) {
            const target = tileAt(x, y);
            if (!canReceiveInfluencePreview(target)) return;

            coords.push({ x, y });
        }

        switch (tile.Structure) {
            case Structure.Settlement:
                addCoord(tile.X, tile.Y);
                break;

            case Structure.Outpost:
                addCoord(tile.X, tile.Y);

                for (const n of hexNeighbors(tile.X, tile.Y)) {
                    addCoord(n.x, n.y);
                }

                break;

            case Structure.City:
                addCoord(tile.X, tile.Y);
                break;

            case Structure.Watchtower:
                addCoord(tile.X, tile.Y);

                for (const n of hexNeighbors(tile.X, tile.Y)) {
                    addCoord(n.x, n.y);
                }

                for (const n of hexRingRadius2(tile.X, tile.Y)) {
                    if (
                        canInfluenceDistance2Preview(tile.X, tile.Y, n.x, n.y)
                    ) {
                        addCoord(n.x, n.y);
                    }
                }

                break;

            case Structure.Bridge:
                addCoord(tile.X, tile.Y);

                for (const n of hexNeighbors(tile.X, tile.Y)) {
                    addCoord(n.x, n.y);
                }

                break;
        }

        return coords;
    }

    function handleStructureHover(tile: Tile | undefined) {
        if (!tile || tile.Structure === Structure.None) {
            hoveredStructureKey = "";
            hoveredStructureInfluence = new Set();
            hoveredStructureOwner = 0;
            return;
        }

        hoveredStructureKey = key(tile.X, tile.Y);
        hoveredStructureOwner = tile.StructureOwner;

        hoveredStructureInfluence = new Set(
            structureInfluenceCoords(tile).map((coord) =>
                key(coord.x, coord.y),
            ),
        );
    }

    function clearStructureHover() {
        hoveredStructureKey = "";
        hoveredStructureInfluence = new Set();
        hoveredStructureOwner = 0;
    }

    function isStructurePreviewed(hex: RenderHex) {
        if (!hex.tile) return false;
        if (!hoveredStructureKey) return false;

        return hoveredStructureInfluence.has(key(hex.tile.X, hex.tile.Y));
    }

    function tileInfluenceRows(tile: Tile | undefined): InfluenceTooltipRow[] {
        if (!tile) return [];

        const rows = game.Players.map((player) => {
            const influence = Number(tile.Influence?.[player.Id] ?? 0);
            const tempInfluence = Number(tile.TempInfluence?.[player.Id] ?? 0);

            return {
                playerId: player.Id,
                influence,
                tempInfluence,
                total: influence + tempInfluence,
                leading: false,
            };
        }).filter((row) => row.total > 0);

        if (rows.length === 0) return [];

        const best = Math.max(...rows.map((row) => row.total));

        return rows
            .map((row) => ({
                ...row,
                leading: row.total === best && best > 0,
            }))
            .sort((a, b) => b.total - a.total);
    }

    function tileTooltipTitle(tile: Tile | undefined, candidate: boolean) {
        if (candidate) return "Place tile";
        if (!tile) return "";

        const biome = biomeLabel(tile.Biome);
        const owner = tile.HasOwner ? `P${tile.Owner}` : "Open";
        const structure =
            tile.Structure !== Structure.None
                ? structureLabel(tile.Structure)
                : "No structure";

        return `${biome} · ${owner}`;
    }

    function tileTooltipSubtitle(tile: Tile | undefined, candidate: boolean) {
        if (candidate) return "Empty placement candidate";
        if (!tile) return "";

        const parts: string[] = [];

        if (tile.Structure !== Structure.None) {
            parts.push(structureLabel(tile.Structure));
        }

        if (tile.HasBlockade) {
            parts.push(`Blockaded by P${tile.BlockadeOwner}`);
        }

        if (tile.Biome === Biome.River && tile.Structure !== Structure.Bridge) {
            parts.push("Unbridged river");
        }

        return parts.join(" · ");
    }
</script>

<section
    class="min-h-[calc(100vh-132px)] min-w-0 overflow-hidden rounded-[34px] bg-[#caa66d] p-4 shadow-[0_18px_0_rgba(44,31,21,0.28)] ring-1 ring-black/20"
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

        <div
            class="h-[calc(100vh-230px)] max-w-full overflow-auto rounded-2xl bg-[#d9c291] p-4 shadow-inner"
        >
            <div
                class="relative"
                style={`width: ${boardSize.width}px; height: ${boardSize.height}px;`}
            >
                {#each renderHexes as hex}
                    {@const targetStatus = getHexTargetStatus(hex)}

                    <BoardTile
                        left={hex.left}
                        top={hex.top}
                        width={HEX_W}
                        height={HEX_H}
                        tile={hex.tile}
                        candidate={hex.candidate}
                        ghost={hex.ghost}
                        clickable={targetStatus.clickable}
                        dimmed={targetStatus.dimmed}
                        biomeClass={biomeClass(hex.tile, hex.candidate)}
                        ownerClass={hex.tile ? ownerClass(hex.tile) : ""}
                        ownerLabel={hex.tile ? ownerLabel(hex.tile) : ""}
                        structureLabel={hex.tile
                            ? structureLabel(hex.tile.Structure)
                            : ""}
                        structureIcon={hex.tile
                            ? structureIcon(hex.tile.Structure)
                            : ""}
                        tooltip={tileTooltip(hex.tile, hex.candidate)}
                        tooltipTitle={tileTooltipTitle(hex.tile, hex.candidate)}
                        tooltipSubtitle={tileTooltipSubtitle(
                            hex.tile,
                            hex.candidate,
                        )}
                        influenceRows={tileInfluenceRows(hex.tile)}
                        auraOwner={tileAuraOwner(hex.tile)}
                        auraEdges={auraEdgesForTile(hex.tile)}
                        influencePreviewed={isStructurePreviewed(hex)}
                        influencePreviewOwner={hoveredStructureOwner}
                        onStructureHover={() => handleStructureHover(hex.tile)}
                        onStructureLeave={clearStructureHover}
                        on:select={() => handleHexSelect(hex)}
                    />
                {/each}
            </div>
        </div>
    </div>
</section>
