import React, { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import "@excalidraw/excalidraw/index.css";

// import type * as TExcalidraw from "@excalidraw/excalidraw";

import App from "./ExampleApp";

declare global {
  interface Window {
    ExcalidrawLib: typeof TExcalidraw;
  }
}

const rootElement = document.getElementById("root")!;
const root = createRoot(rootElement);
console.log('2',window.ExcalidrawLib )
const { Excalidraw } = window.ExcalidrawLib;
root.render(
  <StrictMode>
    <App
      appTitle={"Excalidraw Example"}
      useCustom={(api: any, args?: any[]) => {}}
      excalidrawLib={window.ExcalidrawLib}
    >
      <Excalidraw />
    </App>
  </StrictMode>,
);