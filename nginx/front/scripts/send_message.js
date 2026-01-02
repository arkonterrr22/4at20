const sendMessageBtn = document.querySelector("#send_message")
const msg_text = document.querySelector("#msg_area")
sendMessageBtn.addEventListener("click", async () => {
            try {
                console.log(msg_text.value)
                const data = await window.api.send_message(chatId, msg_text.value);
                console.log("Ответ сервера:", data);
                msg_text.value = "";
                console.log(data.message);
                displayedMessages.push(data.message);
                render_messages();
                scrollToBottom();
            }
            catch (e) {
                alert("Ошибка отправки сообщения");
            }
        }
        );