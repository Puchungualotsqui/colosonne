<script lang="ts">
    import Landing, { type LandingUser } from "./components/Landing.svelte";

    let loading = false;
    let error = "";

    let user: LandingUser = null;

    function createRoom() {
        loading = true;
        error = "";
        console.log("create room");
        setTimeout(() => {
            loading = false;
        }, 500);
    }

    function joinRoom(roomId: string) {
        console.log("join room", roomId);
    }

    function spectateRoom(roomId: string) {
        console.log("spectate room", roomId);
    }

    function login() {
        console.log("login");
    }

    function signUp() {
        console.log("sign up");
    }
</script>

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
