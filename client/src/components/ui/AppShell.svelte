<script lang="ts">
    export let className = "";
</script>

<main
    class={[
        "relative min-h-screen overflow-hidden bg-[#17313a] font-sans text-[#f8efe0]",
        className,
    ].join(" ")}
>
    <div class="pointer-events-none absolute inset-0 bg-[#15323a]">
        <div class="absolute inset-0 bg-board-flat"></div>
        <div class="absolute inset-0 bg-board-texture opacity-[0.16]"></div>
        <div class="absolute inset-x-0 top-0 h-px bg-white/10"></div>
        <div
            class="absolute inset-x-0 bottom-0 h-40 bg-gradient-to-t from-black/18 to-transparent"
        ></div>
    </div>

    <div class="relative z-10">
        <slot />
    </div>
</main>

<style>
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
