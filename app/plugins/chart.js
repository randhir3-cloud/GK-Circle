import {
  Chart,
  Title,
  Tooltip,
  Legend,
  BarElement,
  LineElement,
  PointElement,
  CategoryScale,
  LinearScale,
  Filler,
} from "chart.js";
export default defineNuxtPlugin(() => {
  Chart.register(
    CategoryScale,
    LinearScale,
    BarElement,
    LineElement,
    PointElement,
    Filler,
    Title,
    Tooltip,
    Legend
  );
});
