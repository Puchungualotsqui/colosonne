<script lang="ts">
    import DraftCard from "./DraftCard.svelte";
    import { GamePhase, type DraftItem, type GameState } from "../lib/types";
    import { debugLog } from "../lib/debug";

    export let game: GameState;
    export let playerId = 0;
    export let role: "player" | "spectator" | "" = "";

    export let onPick: (marketIndex: number) => void;

    const MAX_HAND = 3;

    $: me = game.Players.find((p) => p.Id === playerId);
    $: hand = me?.Hand ?? [];
    $: handCount = hand.length;
    $: remainingPicks = Math.max(0, MAX_HAND - handCount);

    $: canPick =
        role === "player" &&
        game.CurrentPhase === GamePhase.Pick &&
        game.CurrentPlayer === playerId &&
        handCount < MAX_HAND;

    function pick(index: number, item: DraftItem) {
        debugLog("market.pick.click", {
            index,
            canPick,
            role,
            playerId,
            currentPlayer: game.CurrentPlayer,
            currentPhase: game.CurrentPhase,
            handCount,
            marketSize: game.Market.length,
            item,
        });

        if (!canPick) return;
        onPick(index);
    }
</script>

<section
    class="rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10"
>
    <div class="flex items-center justify-between gap-3">
        <h2 class="text-xl font-black text-[#fff7e8]">Market</h2>

        <div
            class={[
                "rounded-xl px-3 py-1 text-xs font-black uppercase tracking-wider",
                canPick
                    ? "bg-[#f2c36b] text-[#142833]"
                    : "bg-[#f8efe0]/10 text-[#9fc9c5]",
            ].join(" ")}
        >
            {#if canPick}
                Pick {remainingPicks}
            {:else if game.CurrentPhase === GamePhase.Pick}
                {game.Market.length} left
            {:else}
                {game.Market.length} cards
            {/if}
        </div>
    </div>

    <div class="mt-2 text-xs font-bold text-[#9fc9c5]">
        {#if game.CurrentPhase === GamePhase.Pick}
            {#if canPick}
                Choose up to 3 cards. The market refills for the next player.
            {:else}
                Waiting for P{game.CurrentPlayer} to draft.
            {/if}
        {:else}
            Market refills at the next draft phase.
        {/if}
    </div>

    <div class="mt-4 grid grid-cols-2 gap-3">
        {#each game.Market as item, index (index)}
            <DraftCard
                {item}
                {index}
                disabled={!canPick}
                onPick={(pickedIndex) => pick(pickedIndex, item)}
            />
        {/each}

        {#if game.Market.length === 0}
            <div
                class="col-span-2 rounded-2xl border-2 border-dashed border-[#f8efe0]/20 bg-[#f8efe0]/6 p-5 text-center text-sm font-bold text-[#9fc9c5]"
            >
                Market empty
            </div>
        {/if}
    </div>
</section>
