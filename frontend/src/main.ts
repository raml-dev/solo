import "./style.css";
import { mount } from "svelte";
import App from "./App.svelte";

let target = document.getElementById("app");

if (!target) {
  target = document.createElement("div");
  target.id = "app";
  document.body.appendChild(target);
}
const app = mount(App, { target });

export default app;
