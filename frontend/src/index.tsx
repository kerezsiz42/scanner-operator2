import React from "react";
import ReactDOM from "react-dom/client";
import { Navbar } from "./components/Navbar";
import { Footer } from "./components/Footer";
import { GlobalStateProvider } from "./components/GlobalState";
import { ScanResultList } from "./components/ScanResultList";

function App() {
  return (
    <GlobalStateProvider>
      <div className="h-full flex flex-col">
        <Navbar />
        <ScanResultList />
        <Footer />
      </div>
    </GlobalStateProvider>
  );
}

const element = document.getElementById("app")!;
const root = ReactDOM.createRoot(element);
root.render(<App />);
