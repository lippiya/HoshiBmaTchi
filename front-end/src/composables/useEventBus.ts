import { ref } from "vue";

const events = ref(new Map<string, Function[]>());

export function useEventBus() {
  const emit = (event: string, data?: any) => {
    if (events.value.has(event)) {
      events.value.get(event)?.forEach(callback => {
        callback(data);
      });
    }
  };

  const on = (event: string, callback: Function) => {
    if (!events.value.has(event)) {
      events.value.set(event, []);
    }
    events.value.get(event)?.push(callback);
  };

  const off = (event: string, callback?: Function) => {
    if (!events.value.has(event)) return;

    if (callback) {
      const callbacks = events.value.get(event);
      const index = callbacks?.indexOf(callback);
      if (index && index > -1) {
        callbacks?.splice(index, 1);
      }
    } else {
      events.value.delete(event);
    }
  };

  return {
    emit,
    on,
    off
  };
}