import "$src/app.css";
import App from "$src/App.svelte";
import { mount } from "svelte";

let target = document.getElementById("app");

if (!target) {
  target = document.createElement("div");
  target.id = "app";
  document.body.appendChild(target);
}
const app = mount(App, { target });

export default app;
