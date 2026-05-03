"use client";

import React, { createContext, useContext, useState, ReactNode } from "react";
import ServerError from "@/components/ServerError";

interface ApiContextType {
  isOffline: boolean;
  setOffline: (offline: boolean) => void;
}

const ApiContext = createContext<ApiContextType | undefined>(undefined);

export function ApiProvider({ children }: { children: ReactNode }) {
  const [isOffline, setIsOffline] = useState(false);

  React.useEffect(() => {
    const handleOffline = () => setIsOffline(true);
    window.addEventListener("server-offline", handleOffline);
    return () => window.removeEventListener("server-offline", handleOffline);
  }, []);

  if (isOffline) {
    return <ServerError onRetry={() => setIsOffline(false)} />;
  }

  return (
    <ApiContext.Provider value={{ isOffline, setOffline: setIsOffline }}>
      {children}
    </ApiContext.Provider>
  );
}

export function useApi() {
  const context = useContext(ApiContext);
  if (context === undefined) {
    throw new Error("useApi must be used within an ApiProvider");
  }
  return context;
}
