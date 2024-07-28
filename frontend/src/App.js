import React, { useState } from "react";
import axios from "axios";

function App() {
  const [key, setKey] = useState("");
  const [value, setValue] = useState("");
  const [expiration, setExpiration] = useState("");
  const [retrievedValue, setRetrievedValue] = useState("");

  const handleSet = async () => {
    await axios.post("http://localhost:8080/set", {
      key,
      value,
      expiration: parseInt(expiration),
    });
  };

  const handleGet = async () => {
    try {
      const response = await axios.get(`http://localhost:8080/get/${key}`);
      setRetrievedValue(response.data.value);
    } catch (error) {
      setRetrievedValue("Key not found or expired");
    }
  };

  return (
    <div>
      <h1>LRU Cache</h1>
      <div>
        <input
          type="text"
          placeholder="Key"
          value={key}
          onChange={(e) => setKey(e.target.value)}
        />
        <input
          type="text"
          placeholder="Value"
          value={value}
          onChange={(e) => setValue(e.target.value)}
        />
        <input
          type="text"
          placeholder="Expiration (seconds)"
          value={expiration}
          onChange={(e) => setExpiration(e.target.value)}
        />
        <button onClick={handleSet}>Set</button>
      </div>
      <div>
        <button onClick={handleGet}>Get</button>
        <p>Retrieved Value: {retrievedValue}</p>
      </div>
    </div>
  );
}

export default App;
