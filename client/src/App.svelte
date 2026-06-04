<script lang="ts">
    import Landing, { type LandingUser } from "./components/Landing.svelte";
    import Lobby from "./components/Lobby.svelte";
    import type {
        GameState,
        RoomIdentity,
        RoomState,
        ServerMessage,
    } from "./lib/types";
    import { guestLogin, getMe } from "./lib/api";
    import { GameSocket } from "./lib/ws";

    let user: LandingUser = null;
    let loading = false;
    let error = "";

    let socket: GameSocket | null = null;

    let roomId = "";
    let playerId = 0;
    let role: "player" | "spectator" | "" = "";

    let roomState: RoomState | null = null;
    let game: GameState | null = null;

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

        socket = new GameSocket({
            onMessage: handleMessage,
            onOpen: () => {
                socket?.send(type, data);
            },
            onClose: () => {
                console.log("socket closed");
            },
            onError: () => {
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
        switch (msg.type) {
            case "room_created":
            case "room_joined":
            case "room_spectating":
                applyRoomIdentity(msg.data);
                loading = false;
                break;

            case "room_state":
                roomState = msg.data;
                roomId = msg.data.roomId;
                game = msg.data.game;
                loading = false;
                error = "";
                break;

            case "state":
                // Backward compatibility if your backend still sends "state".
                roomId = msg.data.roomId;
                game = msg.data.game;
                loading = false;
                error = "";
                break;

            case "room_waiting":
                // Backward compatibility if your backend still sends "room_waiting".
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
    <main class="min-h-screen bg-[#15323a] p-6 text-[#f8efe0]">
        <div class="mx-auto max-w-6xl">
            <div class="mb-6 flex items-center justify-between gap-4">
                <div>
                    <div
                        class="text-sm font-bold uppercase tracking-[0.2em] text-[#9fc9c5]"
                    >
                        Room
                    </div>

                    <div class="mt-1 flex items-center gap-3">
                        <div
                            class="rounded-2xl bg-[#f2c36b] px-4 py-2 text-2xl font-black tracking-wider text-[#142833]"
                        >
                            {roomId}
                        </div>

                        <button
                            class="cursor-pointer rounded-xl bg-[#f8efe0]/10 px-3 py-2 text-sm font-bold text-[#fff7e8] ring-1 ring-[#f8efe0]/20 hover:bg-[#f8efe0]/16"
                            type="button"
                            on:click={copyRoomCode}
                        >
                            Copy
                        </button>
                    </div>
                </div>

                <button
                    class="cursor-pointer rounded-xl bg-[#b94b3f] px-4 py-2 font-bold text-white hover:bg-[#c9574a]"
                    type="button"
                    on:click={leaveRoom}
                >
                    Leave
                </button>
            </div>

            <div class="grid gap-4 lg:grid-cols-[280px_1fr]">
                <aside
                    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
                >
                    <div class="text-lg font-bold">Session</div>

                    <div class="mt-4 space-y-2 text-sm text-[#d9e6df]">
                        <div>
                            Role:
                            <span class="font-bold text-[#fff7e8]">
                                {role || "unknown"}
                            </span>
                        </div>

                        <div>
                            Player ID:
                            <span class="font-bold text-[#fff7e8]">
                                {playerId || "-"}
                            </span>
                        </div>

                        <div>
                            Round:
                            <span class="font-bold text-[#fff7e8]">
                                {game.Round}
                            </span>
                        </div>

                        <div>
                            Phase:
                            <span class="font-bold text-[#fff7e8]">
                                {game.CurrentPhase}
                            </span>
                        </div>

                        <div>
                            Current Player:
                            <span class="font-bold text-[#fff7e8]">
                                {game.CurrentPlayer}
                            </span>
                        </div>
                    </div>
                </aside>

                <section
                    class="rounded-3xl bg-[#ead7aa] p-5 text-[#142833] shadow-[0_12px_0_rgba(44,31,21,0.22)] ring-1 ring-black/20"
                >
                    <div class="mb-4 flex items-center justify-between">
                        <div>
                            <div class="text-2xl font-black">
                                Frontiers Match
                            </div>
                            <div class="text-sm font-bold text-[#6b4a2f]">
                                Temporary debug game view
                            </div>
                        </div>
                    </div>

                    <div class="grid gap-4 md:grid-cols-2">
                        <div class="rounded-2xl bg-[#fff7e8] p-4">
                            <div class="font-black">Market</div>

                            <div class="mt-3 grid gap-2">
                                {#each game.Market as item, index}
                                    <button
                                        class="cursor-pointer rounded-xl bg-[#f2c36b] px-3 py-2 text-left text-sm font-bold hover:bg-[#ffd27c] disabled:cursor-not-allowed disabled:opacity-50"
                                        type="button"
                                        disabled={role !== "player" ||
                                            game.CurrentPlayer !== playerId}
                                        on:click={() =>
                                            socket?.send("pick", {
                                                marketIndex: index,
                                            })}
                                    >
                                        [{index}] Kind {item.Kind}
                                    </button>
                                {/each}
                            </div>
                        </div>

                        <div class="rounded-2xl bg-[#fff7e8] p-4">
                            <div class="font-black">Map</div>

                            <div
                                class="mt-3 max-h-[320px] overflow-auto rounded-xl bg-[#ead7aa] p-3 font-mono text-xs"
                            >
                                {#each game.Map as tile}
                                    <div>
                                        ({tile.X}, {tile.Y}) biome={tile.Biome}
                                        owner={tile.HasOwner ? tile.Owner : "-"}
                                        structure={tile.Structure}
                                    </div>
                                {/each}
                            </div>
                        </div>
                    </div>
                </section>
            </div>

            {#if error}
                <div
                    class="mt-5 rounded-2xl bg-[#b94b3f] px-5 py-3 font-semibold text-white"
                >
                    {error}
                </div>
            {/if}
        </div>
    </main>
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
