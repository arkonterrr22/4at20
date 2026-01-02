function searchfn({
    parent,
    placeholder = '',
    data = [],
    displayedProp,
    onSelect = () => {}
}) {
    const input = document.createElement('input');
    input.className = 'searchInput';
    input.placeholder = placeholder;
    input.autocomplete = 'off';

    const dropdown = document.createElement('div');
    dropdown.className = 'searchResults dropdown';
    dropdown.style.display = 'none';

    parent.appendChild(input);
    parent.appendChild(dropdown);

    function showDropdown(items) {
        dropdown.innerHTML = '';

        items.forEach(d => {
            const row = document.createElement('div');
            row.className = 'dropdown-item';
            row.textContent = d[displayedProp];

            row.onclick = () => {
                onSelect(d);
                dropdown.style.display = 'none';
            };

            dropdown.appendChild(row);
        });

        dropdown.style.display = items.length ? 'block' : 'none';
    }

    input.addEventListener('click', () => {
        const filtered = data.filter(d =>
            d[displayedProp]
        );
        showDropdown(filtered);
    });

    input.addEventListener('input', () => {
        const q = input.value.toLowerCase().trim();
        if (!q) {
            dropdown.style.display = 'none';
            return;
        }

        const filtered = data.filter(d =>
            d[displayedProp]
             ?.toLowerCase()
             .includes(q)
        );

        showDropdown(filtered);
    });

    document.addEventListener('click', e => {
        if (!parent.contains(e.target)) {
            dropdown.style.display = 'none';
        }
    });

    return {
        setData(newData) {
            data = newData;
        },
        clear() {
            input.value = '';
            dropdown.style.display = 'none';
        }
    };
}