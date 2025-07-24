const contentEl = document.getElementById("content");

document.addEventListener("keydown", (event) => {
  if (event.ctrlKey && event.key === "a") {
    event.preventDefault();

    const range = document.createRange();
    range.selectNodeContents(contentEl);

    const selection = window.getSelection();
    selection.removeAllRanges();
    selection.addRange(range);
  }
});

hljs.highlightElement(contentEl);
