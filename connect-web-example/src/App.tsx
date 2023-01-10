import { useState } from 'react'
import './App.css'
import {
    createConnectTransport,
    createPromiseClient,
} from "@bufbuild/connect-web";

// Import service definition that you want to connect to.
import { Kv } from "../gen/read_connectweb";

// The transport defines what type of endpoint we're hitting.
// In our example we'll be communicating with a Connect endpoint.
const transport = createConnectTransport({
    baseUrl: "http://localhost:8000",
});

// Here we make the client itself, combining the service
// definition with the transport.
const client = createPromiseClient(Kv, transport);

function App() {
    const [inputValue, setInputValue] = useState("");
    const [messages, setMessages] = useState<
        {
            message: string;
        }[]
    >([]);
    return <>
        <h2>Enter a block ID (without 0x prefix) to get the value from the kv store</h2>
        <ol>
            {messages.map((msg, index) => (
                <li key={index}>
                    {msg.message}
                </li>
            ))}
        </ol>
        <form onSubmit={async (e) => {
            e.preventDefault();
            // Clear inputValue since the user has submitted.
            setInputValue("");
            setMessages((prev) => [
                ...prev,
                {
                    message: "Request: " + inputValue,
                },
            ]);
            try {
                const response = await client.get({
                    key: inputValue,
                });
                setMessages((prev) => [
                    ...prev,
                    {
                        message: "Response: " + bufferToHex(response.value),
                        error: "",
                    },
                ]);
            } catch (e) {
                setMessages((prev) => [
                    ...prev,
                    {
                        message: "Error: " + JSON.stringify(e),
                    },
                ]);

            }
        }}>
            <input value={inputValue} onChange={e => setInputValue(e.target.value)} />
            <button type="submit">Send</button>
        </form>
    </>;
}


function bufferToHex(buffer: Uint8Array): string {
    var s = '', h = '0123456789ABCDEF';
    (new Uint8Array(buffer)).forEach((v) => { s += h[v >> 4] + h[v & 15]; });
    return s;
}

export default App