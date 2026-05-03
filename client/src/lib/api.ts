const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export class ConnectionError extends Error {
  constructor(message: string = "サーバーに接続できません。サーバーが起動しているか確認してください。") {
    super(message);
    this.name = "ConnectionError";
  }
}

export async function apiRequest(path: string, options: RequestInit = {}) {
  const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
  
  const headers = {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  try {
    const response = await fetch(`${API_URL}${path}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      if (response.status === 401) {
        if (typeof window !== "undefined") {
          localStorage.removeItem("token");
          localStorage.removeItem("user");
          window.location.href = "/login";
        }
      }
      let errorMessage = "API request failed";
      const text = await response.text();
      try {
        const errorData = JSON.parse(text);
        errorMessage = errorData.error || errorMessage;
      } catch (e) {
        errorMessage = text || `Error: ${response.status} ${response.statusText}`;
      }
      throw new Error(errorMessage);
    }

    const text = await response.text();
    try {
      return JSON.parse(text);
    } catch (e) {
      return text;
    }
  } catch (error: any) {
    // ネットワークエラー（サーバー停止、DNSエラー、オフライン等）の検知
    // ブラウザによってメッセージが異なるため、代表的なものをチェックするか TypeError 全体を対象にする
    const isNetworkError = 
      error instanceof TypeError && 
      (error.message === "Failed to fetch" || 
       error.message.includes("NetworkError") || 
       error.message.includes("network error") ||
       error.message.includes("Failed to execute 'fetch'"));

    if (isNetworkError) {
      if (typeof window !== "undefined") {
        window.dispatchEvent(new Event("server-offline"));
      }
      throw new ConnectionError();
    }
    throw error;
  }
}
