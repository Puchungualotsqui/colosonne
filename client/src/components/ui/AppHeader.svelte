<script lang="ts">
    export let title = "Frontiers";
    export let subtitle = "";
</script>

<header
    class="mx-auto flex max-w-7xl items-center justify-between px-6 py-5 lg:px-12"
>
    <div class="flex items-center gap-3">
        <div
            class="grid h-11 w-11 place-items-center rounded-2xl bg-[#f2c36b] text-xl font-black text-[#142833] shadow-[0_8px_0_rgba(0,0,0,0.16)] ring-1 ring-white/20"
        >
            <span class="logo-diamond">◈</span>
        </div>

        <div>
            <div class="text-xl font-semibold tracking-tight text-[#fff7e8]">
                {title}
            </div>

            {#if subtitle}
                <div
                    class="text-xs font-semibold uppercase tracking-[0.22em] text-[#9fc9c5]"
                >
                    {subtitle}
                </div>
            {/if}
        </div>
    </div>

    <div class="flex items-center gap-3">
        <slot name="actions" />
    </div>
</header>

<style>
    .logo-diamond {
        display: block;
        line-height: 1;
        transform: translateY(-1px);
    }
</style>
