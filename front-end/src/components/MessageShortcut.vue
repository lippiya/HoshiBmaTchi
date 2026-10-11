<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();

const position = ref({ x: window.innerWidth - 220, y: window.innerHeight - 70 });
const isDragging = ref(false);
const dragStart = ref({ x: 0, y: 0 });
const initialPosition = ref({ x: 0, y: 0 });
const dragThreshold = 5;

const toMessages = () => {
  const distance = Math.sqrt(
    Math.pow(position.value.x - initialPosition.value.x, 2) +
    Math.pow(position.value.y - initialPosition.value.y, 2)
  );
  
  if (distance <= dragThreshold) {
    router.push("/messages");
  }
};

const onMouseDown = (e: MouseEvent) => {
  isDragging.value = true;
  dragStart.value = {
    x: e.clientX - position.value.x,
    y: e.clientY - position.value.y
  };
  initialPosition.value = { ...position.value };
  e.preventDefault();
};

const onMouseMove = (e: MouseEvent) => {
  if (isDragging.value) {
    position.value = {
      x: e.clientX - dragStart.value.x,
      y: e.clientY - dragStart.value.y
    };
  }
};

const onMouseUp = () => {
  isDragging.value = false;
};

if (typeof window !== "undefined") {
  window.addEventListener("mousemove", onMouseMove);
  window.addEventListener("mouseup", onMouseUp);
}

</script>

<template>
  <div
    class="container"
    :style="{ left: position.x + 'px', top: position.y + 'px' }"
    @mousedown="onMouseDown"
    @click="toMessages"
  >
    <h1>Messages</h1>
  </div>
</template>

<style scoped>
.container {
  position: fixed;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 200px;
  height: 50px;
  
  border: 1px solid white;
  background-color: rgba(0, 0, 0, 0.8);
  border-radius: 100px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
  color: white;
  padding: 10px;
  
  z-index: 200;
  user-select: none;
  transition: background-color 0.3s ease;
}

h1{
    font-size: 18px;
    font-weight: 600;
}

.container:hover {
  cursor: grab;
  background-color: rgba(255, 255, 255, 0.1);
}

.container:active {
  cursor: grabbing;
}

</style>