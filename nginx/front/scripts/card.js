function cardfn({
    parent,
    contenthtml = '',
    dataset = {},
    buttons = [{ text: '&times', classname: 'closebtn', fun: () => {} }],
}) {
    const card = document.createElement('div');
    card.className = 'card';

    for (const [key, value] of Object.entries(dataset)) {
        card.dataset[key] = value;
    }

    const content = document.createElement('div');
    content.innerHTML = contenthtml;
    card.appendChild(content);

    buttons.forEach(b => {
        const button = document.createElement('button');
        button.type = "button";
        button.innerText = b.text;
        button.className = b.classname;
        button.addEventListener('click', (e) => b.fun(e));
        card.appendChild(button);
    });

    parent.appendChild(card);
}