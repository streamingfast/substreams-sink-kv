import { useState } from 'react'
import './App.css'
import {
    createConnectTransport,
    createPromiseClient,
    ConnectError,
    codeToString,
} from "@bufbuild/connect-web";

// Import service definition that you want to connect to.
import { Kv } from "../gen/read_connectweb";
import { BlockMeta } from "../gen/block_meta_pb";

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
            color: string;
        }[]
    >([]);
    return <>
        <h1>Example UI for substreams-eth-block-meta</h1>
        <h2>Enter a key to get the value from the kv store (ex: month:last:201511)</h2>
        <ol>
            {messages.map((msg, index) => (
                <li key={index}>
                    <div style={{ color: msg.color }}>
                        <pre>{msg.message}</pre>
                    </div>
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
                    color: "grey",
                },
            ]);
            try {
                const response = await client.get({
                    key: inputValue.replace("0x", ""),
                });
                let output: string;
                try {
                    const blkmeta = BlockMeta.fromBinary(response.value);
                    output = JSON.stringify(blkmeta, (key, value) => {
                        if (key === "hash") {
                            return "0x" + bufferToHex(blkmeta.hash);
                        }
                        if (key === "parentHash") {
                            return "0x" + bufferToHex(blkmeta.parentHash);
                        }
                        return value;
                    }, 2);
                } catch (e) {
                    output = "0x" + bufferToHex(response.value);
                }
                setMessages((prev) => [
                    ...prev,
                    {
                        message: output,
                        color: "lightblue",
                    },
                ]);
            } catch (e) {
                let errorMessage: string;

                if (e instanceof ConnectError) {
                    errorMessage = JSON.stringify({
                        name: e.name,
                        code: codeToString(e.code),
                        message: e.rawMessage,
                    }, null, 2);
                } else {
                    errorMessage = JSON.stringify(e, null, 2);
                }
                setMessages((prev) => [
                    ...prev,
                    {
                        message: "Error: " + errorMessage,
                        color: "pink",
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
    var s = '', h = '0123456789abcdef';
    (new Uint8Array(buffer)).forEach((v) => { s += h[v >> 4] + h[v & 15]; });
    return s;
}

export default App
