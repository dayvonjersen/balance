window.onload = function() {
    let rows = document.body.querySelectorAll("tbody tr");
    let hidden = 0;
    let showmoreText = () => hidden > 0 ? `${hidden} entr${hidden == 1 ? 'y' : 'ies'} hidden. show ${hidden > 10 ? '10' : hidden} more entr${hidden == 1 ?'y' : 'ies'} ▼` : '';
    if(rows.length > 10) {
        rows.forEach((tr, i) => {
            if(rows.length - i > 10) {
                tr.style.display = 'none';
            }
        });

        hidden = rows.length - 10;

        let thead = document.body.querySelector('thead');
        thead.insertAdjacentHTML("beforeend", `<tr><td colspan='4' id='showmore'>${showmoreText()}</td></tr>`)
        let showmore = document.body.querySelector('#showmore');
        showmore.addEventListener("click", () => {
            hidden -= 10;
            if(hidden < 0) {
                hidden = 0;
                showmore.style.display = 'none';
            }
            showmore.innerText = showmoreText();
            rows.forEach((tr, i) => {
                if(i >= hidden) {
                    tr.style.display = 'table-row';
                }
            });
        });
    }
    rows.forEach((tr, i) => {
        tr.addEventListener('mouseover', () => {
            let hoversum = document.body.querySelector('#hoversum');
            if(hoversum) {
                hoversum.parentElement.removeChild(hoversum);
            }
            let sum = 0;
            for(let j = 0; j <= i; j++) {
                sum += parseFloat(rows[j].lastElementChild.innerText);
            }
            tr.insertAdjacentHTML("afterend", `<tr id='hoversum'><td colspan='4'>balance at this point: ${sum.toFixed(2)}</td></tr>`);
        });
    });



}
