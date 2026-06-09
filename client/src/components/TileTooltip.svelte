<script lang="ts" context="module">
    export type InfluenceTooltipRow = {
        playerId: number;
        influence: number;
        tempInfluence: number;
        total: number;
        leading: boolean;
    };
</script>

<script lang="ts">
    export let title = "";
    export let subtitle = "";
    export let influenceRows: InfluenceTooltipRow[] = [];
</script>

<div
    class="w-[260px] rounded-2xl bg-[#142833] p-4 text-[#f8efe0] shadow-2xl ring-1 ring-white/10"
>
    <div class="text-sm font-black">
        {title}
    </div>

    {#if subtitle}
        <div class="mt-1 text-xs font-semibold text-[#9fc9c5]">
            {subtitle}
        </div>
    {/if}

    {#if influenceRows.length > 0}
        <div class="mt-3 space-y-2">
            <div
                class="text-[10px] font-black uppercase tracking-[0.18em] text-[#9fc9c5]"
            >
                Influence
            </div>

            {#each influenceRows as row}
                <div
                    class={[
                        "flex items-center justify-between rounded-xl px-3 py-2 text-xs font-black",
                        row.leading
                            ? "bg-[#f2c36b] text-[#142833]"
                            : "bg-[#f8efe0]/10 text-[#f8efe0]",
                    ].join(" ")}
                >
                    <div>P{row.playerId}</div>

                    <div class="flex items-center gap-2">
                        {#if row.tempInfluence > 0}
                            <span class="text-[10px] opacity-70">
                                {row.influence}+{row.tempInfluence}
                            </span>
                        {/if}

                        <span>{row.total}</span>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>
