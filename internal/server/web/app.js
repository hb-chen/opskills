document.addEventListener('DOMContentLoaded', () => {
    const queryInput = document.getElementById('queryInput');
    const sendBtn = document.getElementById('sendBtn');
    const messagesContainer = document.getElementById('messages');
    const resultTab = document.getElementById('resultTab');
    const resultContent = document.getElementById('resultContent');
    const logsContainer = document.getElementById('logsContainer');
    const statusIndicator = document.getElementById('statusIndicator');
    const statusText = statusIndicator.querySelector('.text');
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');

    // Tab switching helper
    function switchTab(tabId) {
        tabBtns.forEach(b => {
            if (b.dataset.tab === tabId) b.classList.add('active');
            else b.classList.remove('active');
        });

        tabContents.forEach(c => c.classList.remove('active'));
        if (tabId === 'result') {
            resultTab.classList.add('active');
        } else {
            document.getElementById('activitiesContent').classList.add('active');
        }
    }

    // Tab click handlers
    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            switchTab(btn.dataset.tab);
        });
    });

    // Auto-resize textarea
    queryInput.addEventListener('input', function () {
        this.style.height = 'auto';
        this.style.height = (this.scrollHeight) + 'px';
        sendBtn.disabled = this.value.trim() === '';
    });

    // Handle send
    sendBtn.addEventListener('click', handleTask);
    queryInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            if (!sendBtn.disabled) handleTask();
        }
    });

    async function handleTask() {
        const query = queryInput.value.trim();
        if (!query) return;

        // Add user message
        addMessage(query, 'user');
        queryInput.value = '';
        queryInput.style.height = 'auto';
        sendBtn.disabled = true;

        // Set status
        setStatus('正在执行...', true);
        switchTab('activities');
        resultContent.innerHTML = '<div class="placeholder-text">正在初始化任务...</div>';
        logsContainer.innerHTML = '';

        try {
            // Start SSE connection
            const eventSource = new EventSource(`/api/run?query=${encodeURIComponent(query)}`);

            eventSource.onmessage = async (event) => {
                const data = JSON.parse(event.data);

                if (data.type === 'update') {
                    // Update status
                    if (data.step) {
                        setStatus(data.step, true);
                    }
                } else if (data.type === 'log') {
                    // Append log
                    const logEntry = document.createElement('div');
                    logEntry.className = 'log-entry';
                    logEntry.textContent = data.message;
                    logsContainer.appendChild(logEntry);
                    logsContainer.scrollTop = logsContainer.scrollHeight;
                } else if (data.type === 'result') {
                    // Final result
                    displayResult(data);
                    renderMath();
                    highlightCode();
                    setStatus('已完成', false);
                    switchTab('result');
                    if (typeof window.processMermaidBlocks === 'function') {
                        setTimeout(window.processMermaidBlocks, 100);
                    }
                    eventSource.close();
                } else if (data.type === 'error') {
                    addMessage(`错误：${data.message}`, 'system');
                    setStatus('错误', false);
                    eventSource.close();
                }
            };

            eventSource.onerror = (err) => {
                console.error('EventSource failed:', err);
                setStatus('连接丢失', false);
                eventSource.close();
            };

        } catch (error) {
            console.error('Error:', error);
            addMessage('启动任务失败。', 'system');
            setStatus('错误', false);
        }
    }

    function displayResult(data) {
        const state = data.state || {};
        let html = '';

        if (state.error) {
            html = `<div style="color: red; padding: 20px;">
                <h2>执行失败</h2>
                <p>${state.error}</p>
            </div>`;
        } else {
            html = '<div style="padding: 20px;">';

            // Task Info
            if (state.task_id) {
                html += `<p><strong>任务 ID:</strong> ${state.task_id}</p>`;
            }

            // Plan
            if (state.plan && state.plan.steps) {
                html += '<h2>执行计划</h2><ul>';
                state.plan.steps.forEach(step => {
                    html += `<li><strong>步骤 ${step.id}:</strong> ${step.description || step.action}</li>`;
                });
                html += '</ul>';
            }

            // Results
            if (state.results && state.results.length > 0) {
                html += '<h2>执行结果</h2>';
                state.results.forEach((result, idx) => {
                    const status = result.success ? '✅' : '❌';
                    html += `<div style="margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 8px;">
                        <h3>${status} 步骤 ${result.step_id || idx + 1}</h3>`;
                    if (result.output) {
                        html += `<pre style="background: #f5f5f5; padding: 10px; border-radius: 4px; overflow-x: auto;">${escapeHtml(result.output)}</pre>`;
                    }
                    if (result.error) {
                        html += `<p style="color: red;">错误: ${escapeHtml(result.error)}</p>`;
                    }
                    html += '</div>';
                });
            }

            // Final Result
            if (state.final_result) {
                html += '<h2>最终结果</h2>';
                if (state.final_result.summary) {
                    html += `<p>${escapeHtml(state.final_result.summary)}</p>`;
                }
                if (state.final_result.output) {
                    html += `<pre style="background: #f5f5f5; padding: 10px; border-radius: 4px; overflow-x: auto;">${escapeHtml(state.final_result.output)}</pre>`;
                }
            }

            html += '</div>';
        }

        resultContent.innerHTML = html;
    }

    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    function addMessage(text, type) {
        const msgDiv = document.createElement('div');
        msgDiv.className = `message ${type}`;

        let avatarSvg = '';
        if (type === 'user') {
            avatarSvg = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>`;
        } else {
            avatarSvg = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></svg>`;
        }

        msgDiv.innerHTML = `
            <div class="avatar">${avatarSvg}</div>
            <div class="content"><p>${escapeHtml(text)}</p></div>
        `;
        messagesContainer.appendChild(msgDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    function setStatus(text, active) {
        statusText.textContent = text;
        if (active) {
            statusIndicator.classList.add('active');
        } else {
            statusIndicator.classList.remove('active');
        }
    }

    function renderMath() {
        if (!window.katex) return;

        const mathBlocks = resultContent.querySelectorAll('code.language-math, code.language-latex, code.language-tex');
        mathBlocks.forEach(block => {
            const latex = block.textContent;
            const span = document.createElement('div');
            span.className = 'math-display';
            span.style.textAlign = 'center';
            span.style.margin = '1em 0';
            try {
                katex.render(latex, span, { displayMode: true, throwOnError: false });
                if (block.parentElement && block.parentElement.tagName === 'PRE') {
                    block.parentElement.replaceWith(span);
                } else {
                    block.replaceWith(span);
                }
            } catch (e) {
                console.error('KaTeX error:', e);
            }
        });

        if (window.renderMathInElement) {
            renderMathInElement(resultContent, {
                delimiters: [
                    { left: '$$', right: '$$', display: true },
                    { left: '\\[', right: '\\]', display: true },
                    { left: '$', right: '$', display: false },
                    { left: '\\(', right: '\\)', display: false }
                ],
                throwOnError: false
            });
        }
    }

    function highlightCode() {
        if (!window.hljs) return;
        resultContent.querySelectorAll('pre code').forEach((block) => {
            hljs.highlightElement(block);
        });
    }
});

