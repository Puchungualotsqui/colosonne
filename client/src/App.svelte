<script lang="ts">
    import Landing, { type LandingUser } from "./components/Landing.svelte";
    import Lobby from "./components/Lobby.svelte";
    import type {
        BuildAction,
        BuildCostsByPlayer,
        GameEvent,
        GameState,
        RoomIdentity,
        RoomState,
        ScoresByPlayer,
        ServerMessage,
    } from "./lib/types";
    import { guestLogin, getMe } from "./lib/api";
    import { GameSocket } from "./lib/ws";
    import GameView from "./components/GameView.svelte";
    import { debugLog } from "./lib/debug";

    let user: LandingUser = null;
    let loading = false;
    let error = "";

    let socket: GameSocket | null = null;

    let roomId = "";
    let playerId = 0;
    let role: "player" | "spectator" | "" = "";

    let roomState: RoomState | null = null;
    let game: GameState | null = null;
    let buildCosts: BuildCostsByPlayer = {};
    let events: GameEvent[] = [];
    let scores: ScoresByPlayer = {};

    $: inRoom = roomId.length > 0;
    $: inLobby = inRoom && roomState?.status === "lobby";
    $: inGame = inRoom && roomState?.status === "playing" && game !== null;

    async function ensureGuest() {
        const me = await getMe();

        if (me.authenticated) {
            user = me;
            return me;
        }

        const guest = await guestLogin();
        user = guest;
        return guest;
    }

    function connectAndSend(type: string, data: unknown = {}) {
        socket?.close();

        debugLog("ws.connectAndSend", {
            type,
            data,
        });

        socket = new GameSocket({
            onMessage: handleMessage,
            onOpen: () => {
                debugLog("ws.open", {
                    sending: type,
                    data,
                });

                socket?.send(type, data);
            },
            onClose: () => {
                debugLog("ws.close");
            },
            onError: () => {
                debugLog("ws.error");
                error = "Connection error";
                loading = false;
            },
        });
    }

    async function createRoom() {
        try {
            loading = true;
            error = "";

            await ensureGuest();
            connectAndSend("create_room", {});
        } catch (err) {
            error =
                err instanceof Error ? err.message : "Could not create room";
            loading = false;
        }
    }

    async function joinRoom(targetRoomId: string) {
        try {
            loading = true;
            error = "";

            const me = await ensureGuest();

            connectAndSend("join_room", {
                roomId: targetRoomId.trim(),
                name: me.displayName,
            });
        } catch (err) {
            error = err instanceof Error ? err.message : "Could not join room";
            loading = false;
        }
    }

    async function spectateRoom(targetRoomId: string) {
        try {
            loading = true;
            error = "";

            const me = await ensureGuest();

            connectAndSend("spectate_room", {
                roomId: targetRoomId.trim(),
                name: me.displayName,
            });
        } catch (err) {
            error = err instanceof Error ? err.message : "Could not watch room";
            loading = false;
        }
    }

    function handleMessage(msg: ServerMessage) {
        debugLog("ws.message", msg);

        switch (msg.type) {
            case "room_created":
            case "room_joined":
            case "room_spectating":
                applyRoomIdentity(msg.data);
                loading = false;
                break;

            case "room_state":
                debugLog("room.state", {
                    roomId: msg.data.roomId,
                    status: msg.data.status,
                    myPlayerId: playerId,
                    myRole: role,
                    currentPlayer: msg.data.game?.CurrentPlayer,
                    currentPhase: msg.data.game?.CurrentPhase,
                    round: msg.data.game?.Round,
                    players: msg.data.players,
                    buildCosts: msg.data.buildCosts,
                    events: msg.data.events,
                });

                roomState = msg.data;
                roomId = msg.data.roomId;
                game = msg.data.game;
                buildCosts = msg.data.buildCosts ?? {};
                events = msg.data.events ?? [];
                loading = false;
                error = "";
                scores = msg.data.scores ?? {};
                break;

            case "state":
                roomId = msg.data.roomId;
                game = msg.data.game;
                buildCosts = {};
                events = [];
                loading = false;
                error = "";
                scores = {};
                break;

            case "room_waiting":
                roomId = msg.data.roomId;
                game = null;
                loading = false;
                break;

            case "kicked": {
                const message =
                    typeof msg.data === "string" ? msg.data : "You were kicked";

                leaveRoom();
                error = message;
                loading = false;
                break;
            }

            case "error":
                error = msg.data;
                loading = false;
                break;
        }
    }

    function applyRoomIdentity(identity: RoomIdentity) {
        debugLog("room.identity", identity);

        roomId = identity.roomId;
        playerId = identity.playerId;
        role = identity.role;
    }

    function leaveRoom() {
        socket?.close();
        socket = null;

        roomId = "";
        playerId = 0;
        role = "";
        roomState = null;
        game = null;
        buildCosts = {};
        loading = false;
    }

    function copyRoomCode() {
        if (!roomId) return;
        navigator.clipboard.writeText(roomId);
    }

    function login() {
        window.location.href = "/auth/google/start";
    }

    function signUp() {
        window.location.href = "/auth/google/start";
    }

    function sendBuild(action: BuildAction, x: number, y: number) {
        debugLog("app.build.send", {
            action,
            x,
            y,
            roomId,
            playerId,
            role,
            currentPlayer: game?.CurrentPlayer,
            currentPhase: game?.CurrentPhase,
            resources: game?.Players.find((p) => p.Id === playerId)?.Resources,
            activeOutpostCountFrontend: game?.Map.filter(
                (tile) =>
                    tile.Structure === 1 &&
                    tile.StructureOwner === playerId &&
                    tile.HasOwner &&
                    tile.Owner === playerId,
            ).length,
            allOutpostCountFrontend: game?.Map.filter(
                (tile) =>
                    tile.Structure === 1 && tile.StructureOwner === playerId,
            ).length,
        });

        socket?.send("build", { action, x, y });
    }
</script>

{#if !inRoom}
    <Landing
        {user}
        {loading}
        {error}
        onCreateRoom={createRoom}
        onJoinRoom={joinRoom}
        onSpectateRoom={spectateRoom}
        onLogin={login}
        onSignUp={signUp}
    />
{:else if inLobby && roomState}
    <Lobby
        room={roomState}
        {loading}
        {error}
        myPlayerId={playerId}
        myRole={role}
        onReady={(ready) => socket?.send("set_ready", { ready })}
        onStartGame={() => socket?.send("start_game", {})}
        onKickPlayer={(targetPlayerId) =>
            socket?.send("kick_player", { playerId: targetPlayerId })}
        onLeaveRoom={leaveRoom}
        onCopyRoomCode={copyRoomCode}
    />
{:else if inGame && game}
    <GameView
        {game}
        {roomId}
        {playerId}
        {role}
        {error}
        {buildCosts}
        {scores}
        {events}
        onPick={(marketIndex) => socket?.send("pick", { marketIndex })}
        onPlaceTile={(handIndex, x, y) =>
            socket?.send("place_tile", { handIndex, x, y })}
        onUseDraft={(handIndex, x, y, targetPlayerId = 0) =>
            socket?.send("use_draft", {
                handIndex,
                x,
                y,
                targetPlayerId,
            })}
        onDiscardDraft={(handIndex) =>
            socket?.send("discard_draft", { handIndex })}
        onPassPlace={() => socket?.send("pass_place", {})}
        onBuild={sendBuild}
        onPassBuild={() => sendBuild("pass", 0, 0)}
        onLeaveRoom={leaveRoom}
        onCopyRoomCode={copyRoomCode}
    />
{:else}
    <main
        class="grid min-h-screen place-items-center bg-[#15323a] p-6 text-[#f8efe0]"
    >
        <div
            class="rounded-3xl bg-[#23444c] p-6 shadow-md ring-1 ring-[#f8efe0]/10"
        >
            <div class="text-xl font-black">Loading room...</div>

            {#if error}
                <div
                    class="mt-4 rounded-2xl bg-[#b94b3f] px-5 py-3 font-semibold text-white"
                >
                    {error}
                </div>
            {/if}

            <button
                class="mt-5 cursor-pointer rounded-xl bg-[#b94b3f] px-4 py-2 font-bold text-white hover:bg-[#c9574a]"
                type="button"
                on:click={leaveRoom}
            >
                Leave
            </button>
        </div>
    </main>
{/if}
