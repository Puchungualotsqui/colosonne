<script lang="ts">
    import ResourceIcon from "./ResourceIcon.svelte";

    export let wood = 0;
    export let stone = 0;
    export let grain = 0;
    export let relic = 0;
    export let affordable = true;
</script>

<div
    class={[
        "flex flex-wrap items-center justify-center gap-1.5 rounded-xl px-2 py-2",
        affordable ? "bg-white/18" : "bg-black/10 opacity-55",
    ].join(" ")}
>
    {#if wood > 0}
        <ResourceIcon resource="wood" amount={wood} />
    {/if}

    {#if stone > 0}
        <ResourceIcon resource="stone" amount={stone} />
    {/if}

    {#if grain > 0}
        <ResourceIcon resource="grain" amount={grain} />
    {/if}

    {#if relic > 0}
        <ResourceIcon resource="relic" amount={relic} />
    {/if}
</div>
