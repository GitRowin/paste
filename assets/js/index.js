const saveEl = document.getElementById("save");
const contentEl = document.getElementById("content");

saveEl.addEventListener("click", submit);

document.addEventListener("keydown", (event) => {
  if ((event.ctrlKey || event.metaKey) && event.key === "s") {
    event.preventDefault();
    void submit();
  }
});

async function submit() {
  if (saveEl.disabled) return;
  saveEl.disabled = true;

  try {
    const response = await fetch("/save", {
      method: "POST",
      body: contentEl.value,
    });

    if (response.status === 429) {
      showErrorToast("You have been rate limited.");
      saveEl.disabled = false;
      return;
    }

    const { error, id } = await response.json();

    if (error) {
      showErrorToast(error);
      saveEl.disabled = false;
      return;
    }

    window.location = "/" + id;
  } catch (e) {
    console.error(e);
    showErrorToast(e);
    saveEl.disabled = false;
  }
}

function showErrorToast(message) {
  const toast = Toastify({
    text: message,
    position: "center",
    gravity: "bottom",
    duration: 5000,
    style: {
      background: "#b91c1c",
      boxShadow: "none",
    },
    onClick: () => toast.hideToast(),
  });

  toast.showToast();
}
