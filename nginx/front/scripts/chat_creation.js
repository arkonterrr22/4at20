const selectedContainer = document.getElementById('selectedMembers');
const submitBtn = document.querySelector("#submitBtn")
let searchel;
let users = [];

async function loadUsers() {
    const res = await window.api.group_users(window.token.groups[0]);
    return res.users.filter(u => u.id !== window.token.user_id);
}

function addMember(user) {
    cardfn({
        parent: selectedContainer,
        contenthtml: `<span>${user.username}</span>`,
        dataset: { id: user.id },
        buttons: [{
            text: 'x',
            classname: 'closebtn',
            fun: (e) => {
                e.target.parentElement.remove();
                users.push(user)
            }
        }]
    })
    users = users.filter(u => u.id !== user.id)
    searchel.setData(users)
}

(async () => {
    users = await loadUsers();
    searchel = searchfn({
        parent: document.querySelector('.search'),
        placeholder: '@ пользователь',
        data: users,
        displayedProp: 'username',
        onSelect: addMember
    });
})();


submitBtn.addEventListener("click", async () => {
    try {
        const data1 = await window.api.create_chat(document.querySelector("#chat_title").value, "");
        console.log("Ответ сервера:", data1);
        const chatid = data1.chat.id;
        const memberIds = Array.from(selectedContainer.querySelectorAll('.member'))
            .map(el => el.dataset.id);
        const data2 = await window.api.add_chat_members(chatid, memberIds);
        console.log("Ответ сервера:", data2);
        window.location.href = `/chat.html?chatId=${chatid}`;
    }
    catch (e) {
        alert(e)
        console.log(e)
    }
}
);