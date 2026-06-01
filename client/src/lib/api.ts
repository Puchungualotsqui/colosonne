export type MeResponse = {
  authenticated: boolean;
  isGuest: boolean;
  userId?: number;
  displayName: string;
  avatarUrl?: string;
  karma: number;
};

export async function guestLogin(): Promise<MeResponse> {
  const res = await fetch("/auth/guest", {
    method: "POST",
    credentials: "include",
  });

  if (!res.ok) {
    throw new Error("Guest login failed");
  }

  return res.json();
}

export async function getMe(): Promise<MeResponse> {
  const res = await fetch("/me", {
    credentials: "include",
  });

  if (!res.ok) {
    throw new Error("Could not fetch session");
  }

  return res.json();
}
