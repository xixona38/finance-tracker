// Keep the original select for FormData; use a shared, keyboard-friendly menu on screen.
let active = null;
let counter = 0;

export function closeSelect(restoreFocus = false) {
  if (!active) return;
  const { trigger, list, wrapper } = active;
  list.hidden = true;
  trigger.setAttribute('aria-expanded', 'false');
  trigger.removeAttribute('aria-activedescendant');
  wrapper.classList.remove('is-open');
  active = null;
  if (restoreFocus && trigger.isConnected) trigger.focus({ preventScroll: true });
}

document.addEventListener('pointerdown', event => {
  if (active && !active.wrapper.contains(event.target)) closeSelect();
});
document.addEventListener('focusin', event => {
  if (active && !active.wrapper.contains(event.target)) closeSelect();
});
window.addEventListener('resize', () => closeSelect());

export function enhanceSelects(root) {
  for (const select of root.querySelectorAll('select:not([data-enhanced])')) {
    select.dataset.enhanced = 'true';
    const label = select.getAttribute('aria-label') || [...(select.closest('label')?.childNodes || [])]
      .filter(node => node.nodeType === 3).map(node => node.textContent).join(' ').trim() || 'Выберите значение';
    const wrapper = document.createElement('div');
    wrapper.className = `custom-select${select.classList.contains('filter-select') ? ' compact-select' : ''}`;
    const trigger = document.createElement('button');
    trigger.type = 'button';
    trigger.className = 'select-trigger';
    trigger.setAttribute('role', 'combobox');
    trigger.setAttribute('aria-haspopup', 'listbox');
    trigger.setAttribute('aria-expanded', 'false');
    trigger.setAttribute('aria-autocomplete', 'none');
    const value = document.createElement('span');
    value.className = 'select-value';
    const arrow = document.createElement('span');
    arrow.className = 'select-arrow';
    arrow.setAttribute('aria-hidden', 'true');
    trigger.append(value, arrow);
    const list = document.createElement('div');
    list.id = `select-list-${++counter}`;
    list.className = 'select-list';
    list.hidden = true;
    list.setAttribute('role', 'listbox');
    list.setAttribute('aria-label', label);
    trigger.setAttribute('aria-controls', list.id);
    const options = [...select.options];
    let highlighted = 0;
    let search = '';
    let lastSearch = 0;
    const items = options.map((option, index) => {
      const item = document.createElement('div');
      item.id = `${list.id}-${index}`;
      item.className = 'select-option';
      item.setAttribute('role', 'option');
      item.setAttribute('aria-disabled', String(option.disabled));
      const text = document.createElement('span');
      text.textContent = option.textContent;
      const check = document.createElement('span');
      check.className = 'select-check';
      check.textContent = '✓';
      check.setAttribute('aria-hidden', 'true');
      item.append(text, check);
      // Focus stays on the combobox, including when choosing with a mouse.
      item.addEventListener('pointerdown', event => event.preventDefault());
      item.addEventListener('click', event => {
        event.preventDefault(); event.stopPropagation(); choose(index);
      });
      list.append(item);
      return item;
    });
    wrapper.append(trigger, list);
    select.after(wrapper);
    select.hidden = true;

    function sync() {
      value.textContent = options[select.selectedIndex]?.textContent || 'Выберите значение';
      trigger.setAttribute('aria-label', `${label}: ${value.textContent}`);
      trigger.disabled = select.disabled;
      items.forEach((item, index) => item.setAttribute('aria-selected', String(index === select.selectedIndex)));
    }
    function highlight(index) {
      if (!items[index] || options[index].disabled) return;
      highlighted = index;
      items.forEach((item, i) => item.classList.toggle('highlighted', i === index));
      trigger.setAttribute('aria-activedescendant', items[index].id);
      // Scroll only the menu, not the surrounding page or modal.
      const item = items[index];
      if (item.offsetTop < list.scrollTop) list.scrollTop = item.offsetTop;
      else if (item.offsetTop + item.offsetHeight > list.scrollTop + list.clientHeight) {
        list.scrollTop = item.offsetTop + item.offsetHeight - list.clientHeight;
      }
    }
    function open() {
      if (trigger.disabled || !options.length) return;
      closeSelect();
      active = { wrapper, trigger, list };
      wrapper.classList.add('is-open');
      const rect = trigger.getBoundingClientRect();
      const modal = select.closest('dialog')?.getBoundingClientRect();
      const below = Math.min(window.innerHeight, modal?.bottom ?? window.innerHeight) - rect.bottom - 16;
      const above = rect.top - Math.max(0, modal?.top ?? 0) - 16;
      const upwards = below < 180 && above > below;
      wrapper.classList.toggle('opens-up', upwards);
      list.style.maxHeight = `${Math.max(60, Math.min(240, (upwards ? above : below) - 8))}px`;
      list.hidden = false;
      trigger.setAttribute('aria-expanded', 'true');
      const initial = select.selectedIndex >= 0 && !options[select.selectedIndex].disabled
        ? select.selectedIndex : options.findIndex(option => !option.disabled);
      highlight(initial);
    }
    function choose(index) {
      if (!options[index] || options[index].disabled || trigger.disabled) return;
      const previous = select.value;
      select.selectedIndex = index;
      sync();
      closeSelect(true);
      if (previous !== select.value) select.dispatchEvent(new Event('change', { bubbles: true }));
    }
    trigger.addEventListener('click', event => {
      event.preventDefault();
      if (active?.trigger === trigger) closeSelect(); else open();
    });
    trigger.addEventListener('keydown', event => {
      const isOpen = active?.trigger === trigger;
      if (event.key === 'Escape' && isOpen) {
        event.preventDefault(); event.stopPropagation(); closeSelect(true); return;
      }
      if (event.key === 'Tab') { closeSelect(); return; }
      if (['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) {
        event.preventDefault();
        if (!isOpen) open();
        if (!isOpen && ['ArrowDown', 'ArrowUp'].includes(event.key)) return;
        const step = event.key === 'ArrowUp' || event.key === 'End' ? -1 : 1;
        let index = event.key === 'Home' ? 0 : event.key === 'End' ? options.length - 1 : highlighted + step;
        while (options[index]?.disabled) index += step;
        highlight(index);
        return;
      }
      if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault();
        if (isOpen) choose(highlighted); else open();
        return;
      }
      if (event.key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
        event.preventDefault();
        if (!isOpen) open();
        const now = Date.now();
        search = (now - lastSearch < 700 ? search : '') + event.key.toLocaleLowerCase('ru');
        lastSearch = now;
        const index = options.findIndex(option => !option.disabled && option.textContent.toLocaleLowerCase('ru').startsWith(search));
        if (index >= 0) highlight(index);
      }
    });
    select.addEventListener('change', sync);
    sync();
  }
}
