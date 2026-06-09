export const ui = {
  button: {
    primary:
      "cursor-pointer rounded-2xl bg-[#c96f3d] px-7 py-4 font-bold text-white shadow-[0_8px_0_rgba(91,48,28,0.55)] transition hover:-translate-y-0.5 hover:bg-[#dc7b45] active:translate-y-1 active:shadow-[0_4px_0_rgba(91,48,28,0.55)] disabled:cursor-not-allowed disabled:opacity-60",

    secondary:
      "cursor-pointer rounded-2xl bg-[#f8efe0]/10 px-7 py-4 font-bold text-[#fff7e8] shadow-[0_8px_0_rgba(0,0,0,0.18)] ring-1 ring-[#f8efe0]/20 transition hover:-translate-y-0.5 hover:bg-[#f8efe0]/16 active:translate-y-1 disabled:cursor-not-allowed disabled:opacity-60",

    smallSecondary:
      "cursor-pointer rounded-xl bg-[#f8efe0]/10 px-3 py-2 text-sm font-bold text-[#fff7e8] ring-1 ring-[#f8efe0]/20 hover:bg-[#f8efe0]/16",

    danger:
      "cursor-pointer rounded-xl bg-[#b94b3f] px-4 py-2 text-sm font-bold text-white shadow-sm transition hover:bg-[#c9574a]",

    cream:
      "cursor-pointer rounded-xl bg-[#f8efe0] px-4 py-2 text-sm font-bold text-[#142833] shadow-sm transition hover:bg-white",

    ghost:
      "cursor-pointer rounded-xl px-4 py-2 text-sm font-semibold text-[#f8efe0] transition hover:bg-white/10",
  },

  panel: {
    dark: "rounded-3xl bg-[#23444c] p-5 shadow-md ring-1 ring-[#f8efe0]/10",
    boardOuter:
      "rounded-[34px] bg-[#caa66d] p-4 shadow-[0_18px_0_rgba(44,31,21,0.28)] ring-1 ring-black/20",
    boardInner:
      "rounded-[26px] border border-[#6b4a2f]/35 bg-[#ead7aa] text-[#142833] shadow-inner",
    soft: "rounded-2xl bg-[#f8efe0]/10 ring-1 ring-[#f8efe0]/10",
  },

  badge: {
    teal: "rounded-xl bg-[#f8efe0]/10 px-3 py-1 text-sm font-bold text-[#9fc9c5]",
    gold: "rounded-xl bg-[#f2c36b] px-3 py-1 text-sm font-black text-[#142833]",
    danger:
      "rounded-xl bg-[#b94b3f]/80 px-3 py-1 text-sm font-black text-white",
  },
};
