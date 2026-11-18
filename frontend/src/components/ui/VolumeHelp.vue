<template>
  <details class="volume-help">
    <summary>Как считается объём</summary>
    <p>
      Складываем вес × повторения во всех рабочих подходах. Например: 20 кг × 10 повторений × 3
      подхода = 600 кг.
    </p>
    <p>
      Это суммарная нагрузка за указанное время, а не вес одного снаряда. Разминка не учитывается;
      подходы без дополнительного веса дают 0 кг.
    </p>
  </details>
</template>

<style scoped>
.volume-help {
  margin-top: 16px;
  color: var(--sportik-text-soft);
  font-size: 0.88rem;
  line-height: 1.6;
}
.volume-help summary {
  width: fit-content;
  max-width: 100%;
  cursor: pointer;
  color: var(--sportik-text);
  font-weight: 600;
  padding: 8px 0;
}
.volume-help p {
  max-width: 68ch;
  margin: 8px 0 0;
}
</style>
