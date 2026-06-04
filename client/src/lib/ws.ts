import type { ServerMessage } from "./types";

export type WSHandlers = {
  onMessage: (msg: ServerMessage) => void;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: (err: Event) => void;
};

export class GameSocket {
  private ws: WebSocket;
  private opened = false;
  private queue: Array<{ type: string; data: unknown }> = [];

  constructor(handlers: WSHandlers) {
    const protocol = location.protocol === "https:" ? "wss:" : "ws:";
    this.ws = new WebSocket(`${protocol}//${location.host}/ws`);

    this.ws.onopen = () => {
      this.opened = true;

      for (const msg of this.queue) {
        this.send(msg.type, msg.data);
      }

      this.queue = [];
      handlers.onOpen?.();
    };

    this.ws.onclose = () => handlers.onClose?.();
    this.ws.onerror = (err) => handlers.onError?.(err);

    this.ws.onmessage = (event) => {
      const msg = JSON.parse(event.data) as ServerMessage;
      handlers.onMessage(msg);
    };
  }

  send(type: string, data: unknown = {}) {
    if (!this.opened) {
      this.queue.push({ type, data });
      return;
    }

    this.ws.send(JSON.stringify({ type, data }));
  }

  close() {
    this.ws.close();
  }
}
