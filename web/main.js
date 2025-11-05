const btncreate = document.getElementById("create");
const btnjoin = document.getElementById("join");

let socket;

const join = async (gameId, playerId) => {
  socket = new WebSocket(
    `ws://localhost:8000/ws?gameId=${gameId}&playerId=${playerId}`,
  );
  socket.addEventListener("open", () => {
    console.log("Connected to the server");
  });
  socket.addEventListener("close", () => {
    console.log("Connetion closed");
  });
  socket.addEventListener("error", () => {
    console.log("Connetion error");
  });
  socket.addEventListener("message", (msg) => {
    console.log(msg.data);
    log(msg.data.action);
  });
};

btncreate.onclick = async () => {
  const res = await fetch("http://localhost:8000/games", {
    method: "POST",
  });
  const body = await res.json();
  console.log(body);
  join(body?.gameId, body?.playerId);
};

btnjoin.onclick = async () => {
  const input = document.getElementById("joininput");
  const res = await fetch(`http://localhost:8000/games/${input.value}/join`, {
    method: "POST",
  });
  const body = await res.json();
  console.log(body);
  join(body?.gameId, body?.playerId);
};

// Enviar mensaje al servidor
document.getElementById("enviar").addEventListener("click", () => {
  const msg = document.getElementById("mensaje").value;
  if (socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type: "message", action: msg }));
    log("📤 Enviado: " + msg);
  } else {
    log("⛔ No se puede enviar, conexión no abierta");
  }
});

// Función helper para mostrar logs
function log(msg) {
  const logEl = document.getElementById("log");
  logEl.textContent += msg + "\n";
}
