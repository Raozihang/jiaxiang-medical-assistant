import ReactDOM from "react-dom/client";
import "antd/dist/reset.css";
import { App } from "@/app/App";
import "@/styles/global.css";

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("根元素未找到");
}

ReactDOM.createRoot(rootElement).render(<App />);
