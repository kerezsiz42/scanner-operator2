//@ts-check

/**
 * @typedef {{signal?: AbortSignal}} SubscriberOptions
 */

export class Subscriber extends EventTarget {
  /** @type {WebSocket | undefined} */
  #ws = undefined;
  /** @type {number | undefined} */
  #timeoutId = undefined;
  #isConnected = false;
  #previousIsConnected = false;
  #shouldBeOpen = true;

  /**
   * @param {string} url
   * @param {SubscriberOptions} [options]
   */
  constructor(url, options) {
    super();
    this.#ws = this.#connect(url);

    options?.signal?.addEventListener("abort", () => {
      this.#shouldBeOpen = false;
      this.#ws?.close();
    });
  }

  /**
   * @param {boolean} newState
   */
  #setState(newState) {
    this.#previousIsConnected = this.#isConnected;
    this.#isConnected = newState;

    if (this.#previousIsConnected !== this.#isConnected) {
      const ce = new CustomEvent("connection", { detail: this.#isConnected });
      this.dispatchEvent(ce);
    }
  }

  /**
   * @param {string} url
   * @returns {WebSocket}
   */
  #connect(url) {
    clearTimeout(this.#timeoutId);
    const ws = new WebSocket(url);

    ws.onopen = () => {
      this.#setState(true);
    };

    /**
     * @param {MessageEvent<string>} ev
     */
    ws.onmessage = (ev) => {
      const detail = JSON.parse(ev.data);
      const ce = new CustomEvent("message", { detail });
      this.dispatchEvent(ce);
    };

    ws.onclose = (_ev) => {
      this.#setState(false);
      if (this.#shouldBeOpen) {
        this.#timeoutId = setTimeout(() => this.#connect(url), 5000);
      }
    };

    return (this.#ws = ws);
  }
}
