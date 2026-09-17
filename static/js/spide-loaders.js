/* =========================================================
   Spide POS — Shared Data Loaders
   Include on every page: <script src="/static/js/spide-loaders.js"></script>
   Depends on: spide-modals.js (for getCurrentUserRole fallback, showNotification)
   Populates window.* for all modal data loads.
   ========================================================= */

(function () {
    'use strict';

    // =========================================================
    // HELPERS (safe to call even if page already defines them)
    // =========================================================
    if (typeof window.getCookie !== 'function') {
        window.getCookie = function (name) {
            const value = `; ${document.cookie}`;
            const parts = value.split(`; ${name}=`);
            if (parts.length === 2) return parts.pop().split(';').shift();
            return null;
        };
    }

    if (typeof window.getCurrentUserRole !== 'function') {
        window.getCurrentUserRole = function () {
            const token = window.getCookie('spide_token');
            if (token) {
                try {
                    const payload = JSON.parse(atob(token.split('.')[1]));
                    return payload.role || 'cashier';
                } catch (e) { }
            }
            try {
                const u = JSON.parse(localStorage.getItem('spide_user') || '{}');
                return u.role || 'cashier';
            } catch (e) { return 'cashier'; }
        };
    }

    if (typeof window.showNotification !== 'function') {
        window.showNotification = function (message, type) {
            const notif = document.createElement('div');
            const bg = type === 'warning' ? 'bg-amber-600' : 'bg-emerald-600';
            notif.className = `fixed top-20 right-4 ${bg} text-white px-6 py-3 rounded-lg shadow-lg z-50 transition-all duration-500`;
            notif.textContent = (type === 'warning' ? '⚠️ ' : '✅ ') + message;
            document.body.appendChild(notif);
            setTimeout(() => { notif.style.opacity = '0'; setTimeout(() => notif.remove(), 500); }, 3000);
        };
    }

    if (typeof window.getShopIdFromURL !== 'function') {
        window.getShopIdFromURL = function () {
            const urlParams = new URLSearchParams(window.location.search);
            const shopId = urlParams.get('shop_id');
            if (shopId) {
                localStorage.setItem('currentShopId', shopId);
                return parseInt(shopId);
            }
            return parseInt(localStorage.getItem('currentShopId')) || 1;
        };
    }

    if (typeof window.getCurrentShopId !== 'function') {
        // Safe accessor for the current shop, wherever it lives
        window.getCurrentShopId = function () {
            if (window.currentShop && window.currentShop.id) return window.currentShop.id;
            return window.getShopIdFromURL();
        };
    }

    window.__receiptWindow = null;

    function getReceiptWindow() {
        if (!window.__receiptWindow || window.__receiptWindow.closed) {
            window.__receiptWindow = window.open('', '_blank', 'width=350,height=600');
        }
        return window.__receiptWindow;
    }

    // =========================================================
    // RECENT SALES
    // =========================================================
    window.loadRecentSales = async function () {
        const container = document.getElementById('recent-sales-list');
        if (!container) return;
        container.innerHTML = '<p class="text-gray-500 text-center py-6">Loading...</p>';
        try {
            const token = window.getCookie('spide_token');
            const res = await fetch('/api/sales/recent', {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            const data = await res.json();
            if (!data || data.length === 0) {
                container.innerHTML = '<p class="text-gray-500 text-center py-6">No sales recorded yet.</p>';
                return;
            }
            let html = '';
            data.forEach(sale => {
                const itemsCount = sale.items ? sale.items.length : 0;
                const paymentType = sale.payment_type || 'cash';
                const paymentColor = paymentType === 'mpesa' ? 'green' : paymentType === 'split' ? 'purple' : 'blue';
                html += `<div class="border rounded-lg p-3 flex justify-between items-center hover:bg-gray-50 transition">
                    <div>
                        <div class="font-bold text-gray-800">Sale #${sale.id}</div>
                        <div class="text-xs text-gray-500">${sale.created_at || ''} • ${itemsCount} item${itemsCount !== 1 ? 's' : ''}</div>
                        <span class="text-xs px-2 py-0.5 rounded bg-${paymentColor}-100 text-${paymentColor}-800">${paymentType.toUpperCase()}</span>
                    </div>
                    <div class="text-right">
                        <div class="font-bold text-purple-900 text-lg">KES ${(sale.total_amount || 0).toFixed(2)}</div>
                        <button onclick="reprintReceipt(${sale.id})" class="mt-1 bg-gray-100 hover:bg-purple-600 hover:text-white text-gray-700 text-xs font-semibold px-3 py-1.5 rounded-lg border transition">🖨️ Reprint</button>
                    </div>
                </div>`;
            });
            container.innerHTML = html;
        } catch (e) {
            container.innerHTML = '<p class="text-red-500 text-center py-6">Error loading sales</p>';
        }
    };

    window.loadProductSalesShops = async function () {
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/shops', {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!response.ok) return;
            const shopList = await response.json();
            const select = document.getElementById('productsales-shop-filter');
            if (!select) return;
            const currentValue = select.value;
            select.innerHTML = '<option value="0">All Shops</option>';
            shopList.forEach(s => {
                const opt = document.createElement('option');
                opt.value = s.id;
                opt.textContent = s.name;
                select.appendChild(opt);
            });
            if (currentValue) select.value = currentValue;
        } catch (e) {
            console.error('Error loading shops:', e);
        }
    };

    window.loadProductSalesReport = async function () {
        const tbody = document.getElementById('productsales-table-body');
        if (!tbody) return;

        const startEl = document.getElementById('productsales-start-date');
        const endEl = document.getElementById('productsales-end-date');
        const shopEl = document.getElementById('productsales-shop-filter');
        const catEl = document.getElementById('productsales-category-filter');
        const prodEl = document.getElementById('productsales-product-filter');

        const startDate = startEl ? startEl.value : new Date().toISOString().split('T')[0];
        const endDate = endEl ? endEl.value : new Date().toISOString().split('T')[0];
        const shopId = shopEl ? (shopEl.value || 0) : 0;
        const category = catEl ? catEl.value : '';
        const productId = prodEl ? prodEl.value : '';

        tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-gray-500">Loading...</td></tr>';

        try {
            const token = window.getCookie('spide_token');
            const params = new URLSearchParams({
                start_date: startDate,
                end_date: endDate,
                shop_id: String(shopId)
            });
            if (category) params.set('category', category);
            if (productId) params.set('product_id', productId);

            const url = '/api/reports/product-sales?' + params.toString();
            const res = await fetch(url, {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!res.ok) {
                const errText = await res.text();
                throw new Error(errText || 'Failed: ' + res.status);
            }
            const data = await res.json();
            window.renderProductSalesReport(data);
        } catch (err) {
            console.error('Product sales error:', err);
            tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-red-500">Error: ' + err.message + '</td></tr>';
        }
    };

    window.renderProductSalesReport = function (data) {
        const tbody = document.getElementById('productsales-table-body');
        if (!tbody) return;

        // API returns a flat array, not {products, totals}
        const products = Array.isArray(data) ? data : (data.products || []);

        // Compute summary totals from the rows
        let totalRevenue = 0;
        let totalCogs = 0;
        let totalProfit = 0;
        let totalUnits = 0;
        products.forEach(p => {
            totalRevenue += p.total_revenue || 0;
            totalCogs += p.total_cost || 0;
            totalProfit += p.net_profit || 0;
            totalUnits += p.units_sold || 0;
        });
        const avgMargin = totalRevenue > 0 ? (totalProfit / totalRevenue) * 100 : 0;

        const setText = (id, text) => { const el = document.getElementById(id); if (el) el.textContent = text; };
        setText('ps-total-revenue', 'KES ' + totalRevenue.toFixed(2));
        setText('ps-total-cogs', 'KES ' + totalCogs.toFixed(2));
        setText('ps-total-profit', 'KES ' + totalProfit.toFixed(2));
        setText('ps-avg-margin', avgMargin.toFixed(1) + '%');

        if (products.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-gray-500">No sales in this range</td></tr>';
            return;
        }

        let html = '';
        products.forEach(p => {
            const margin = p.margin_pct || 0;
            const marginColor = margin > 30 ? 'text-emerald-600' : margin > 15 ? 'text-blue-600' : 'text-amber-600';
            html += `<tr class="border-b hover:bg-blue-50/50 transition">
            <td class="p-2 font-medium text-gray-800">${p.product_name}</td>
            <td class="p-2 text-gray-500">${p.category || 'General'}</td>
            <td class="p-2 text-center font-bold">${p.units_sold}</td>
            <td class="p-2 text-right font-mono">KES ${(p.total_revenue || 0).toFixed(2)}</td>
            <td class="p-2 text-right font-mono text-amber-700">KES ${(p.total_cost || 0).toFixed(2)}</td>
            <td class="p-2 text-right font-mono text-emerald-700">KES ${(p.net_profit || 0).toFixed(2)}</td>
            <td class="p-2 text-right font-mono ${marginColor} font-bold">${margin.toFixed(1)}%</td>
        </tr>`;
        });
        tbody.innerHTML = html;
    };

    window.printProductSalesReport = function () {
        const summary = document.getElementById('productsales-summary');
        const table = document.querySelector('#product-sales-modal table');
        const modal = document.getElementById('product-sales-modal');

        if (!summary || !table) {
            alert('Nothing to print — generate the report first.');
            return;
        }

        // Pull the date range from the form so the printout says what period it covers
        const startEl = document.getElementById('productsales-start-date');
        const endEl = document.getElementById('productsales-end-date');
        const startDate = startEl ? startEl.value : '';
        const endDate = endEl ? endEl.value : '';

        // Pull the company name from localStorage (set at login)
        const companyName = localStorage.getItem('company_name') || 'Spide POS';

        // Shop filter label (only shown if the filter exists and a specific shop is chosen)
        let shopLabel = '';
        const shopEl = document.getElementById('productsales-shop-filter');
        if (shopEl && shopEl.value && shopEl.value !== '0') {
            const selected = shopEl.options[shopEl.selectedIndex];
            if (selected) shopLabel = selected.textContent;
        }

        const printWindow = window.open('', '_blank', 'width=900,height=700');
        if (!printWindow) {
            alert('Pop-up blocked. Please allow pop-ups for this site to print.');
            return;
        }

        const printedAt = new Date().toLocaleString('en-KE', {
            dateStyle: 'short',
            timeStyle: 'short'
        });

        const html = `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Product Sales Report</title>
            <style>
                @page { margin: 15mm; }
                body {
                    font-family: Arial, Helvetica, sans-serif;
                    color: #111;
                    font-size: 12px;
                    line-height: 1.4;
                }
                h1 { font-size: 18px; margin: 0 0 4px 0; }
                .meta { font-size: 11px; color: #555; margin-bottom: 12px; }
                .meta div { margin: 2px 0; }
                .summary {
                    display: grid;
                    grid-template-columns: repeat(4, 1fr);
                    gap: 8px;
                    margin: 12px 0 16px 0;
                }
                .summary .card {
                    border: 1px solid #ddd;
                    border-radius: 6px;
                    padding: 10px;
                    text-align: center;
                }
                .summary .card .label {
                    font-size: 10px;
                    text-transform: uppercase;
                    color: #666;
                    letter-spacing: 0.5px;
                }
                .summary .card .value {
                    font-size: 15px;
                    font-weight: bold;
                    margin-top: 4px;
                }
                table { width: 100%; border-collapse: collapse; font-size: 11px; }
                th, td {
                    padding: 6px 8px;
                    border-bottom: 1px solid #eee;
                    text-align: left;
                }
                th {
                    background: #f3f4f6;
                    font-size: 10px;
                    text-transform: uppercase;
                    letter-spacing: 0.5px;
                    color: #555;
                }
                td.num { text-align: right; font-family: "Courier New", monospace; }
                tr:last-child td { border-bottom: none; }
                .footer {
                    margin-top: 20px;
                    padding-top: 10px;
                    border-top: 1px solid #ddd;
                    font-size: 10px;
                    color: #888;
                    text-align: center;
                }
                @media print {
                    .no-print { display: none; }
                }
            </style>
        </head>
        <body>
            <h1>📈 Product Sales &amp; COGS Report</h1>
            <div class="meta">
                <div><strong>Company:</strong> ${companyName}</div>
                <div><strong>Period:</strong> ${startDate || '—'} to ${endDate || '—'}</div>
                ${shopLabel ? `<div><strong>Shop:</strong> ${shopLabel}</div>` : '<div><strong>Shop:</strong> All Shops</div>'}
                <div><strong>Printed:</strong> ${printedAt}</div>
            </div>

            ${summary.outerHTML.replace(/class="[^"]*"/g, '')}

            <table>
                ${table.innerHTML}
            </table>

            <div class="footer">
                Generated by Spide POS
            </div>

            <script>
                window.onload = function () {
                    window.print();
                    setTimeout(function () { window.close(); }, 500);
                };
            <\/script>
        </body>
        </html>
    `;

        printWindow.document.write(html);
        printWindow.document.close();
    };

    window.loadProductSalesCategories = async function () {
        const select = document.getElementById('productsales-category-filter');
        if (!select) return;

        // Preserve current selection
        const currentValue = select.value;

        try {
            const token = window.getCookie('spide_token');
            const res = await fetch('/api/categories', {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!res.ok) return;
            const categories = await res.json();

            select.innerHTML = '<option value="">All Categories</option>';
            (categories || []).forEach(cat => {
                const opt = document.createElement('option');
                opt.value = cat.name || cat;
                opt.textContent = cat.name || cat;
                select.appendChild(opt);
            });

            if (currentValue) select.value = currentValue;
        } catch (e) {
            console.error('Error loading categories:', e);
        }
    };

    window.loadProductSalesProducts = async function () {
        const select = document.getElementById('productsales-product-filter');
        const categorySelect = document.getElementById('productsales-category-filter');
        if (!select) return;

        const category = categorySelect ? categorySelect.value : '';
        const currentValue = select.value;

        try {
            const token = window.getCookie('spide_token');
            // Fetch all products; filter client-side by category
            const res = await fetch('/api/products?shop_id=0', {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!res.ok) return;
            const products = await res.json();

            select.innerHTML = '<option value="">All Products</option>';
            (products || [])
                .filter(p => !category || (p.category || '').toLowerCase() === category.toLowerCase())
                .sort((a, b) => (a.name || '').localeCompare(b.name || ''))
                .forEach(p => {
                    const opt = document.createElement('option');
                    opt.value = p.id;
                    opt.textContent = p.name + (p.category ? ' — ' + p.category : '');
                    select.appendChild(opt);
                });

            // Preserve selection if the product is still in the list
            if (currentValue) {
                const stillExists = [...select.options].some(o => o.value === currentValue);
                if (stillExists) select.value = currentValue;
            }

        } catch (e) {
            console.error('Error loading products:', e);
        }
    };

    window.onProductSalesCategoryChange = async function () {
        await window.loadProductSalesProducts();
        if (typeof window.loadProductSalesReport === 'function') {
            window.loadProductSalesReport();
        }
    };

    window.reprintReceipt = function (saleId) {
        fetch('/api/sales/recent')
            .then(res => res.json())
            .then(data => {
                const sale = data.find(s => s.id === saleId);
                if (!sale) { alert('Sale not found!'); return; }
                const items = (sale.items || []).map(item => ({
                    product_id: item.product_id,
                    product_name: item.product_name || 'Product',
                    quantity: item.quantity,
                    unit_price: item.unit_price,
                    subtotal: item.subtotal
                }));
                const saleData = {
                    total_amount: sale.total_amount,
                    cash_amount: sale.cash_amount || 0,
                    mpesa_amount: sale.mpesa_amount || 0,
                    deposit_amount: sale.deposit_amount || 0,
                    credit_amount: sale.credit_amount || 0,
                    customer_id: sale.customer_id || 0,
                    mpesa_code: sale.mpesa_code || '',
                    change_given: 0,
                    items: items
                };
                window.printThermalReceipt(saleData, saleId, true);
            })
            .catch(err => { console.error(err); alert('Error reprinting.'); });
    };

    window.printThermalReceipt = function (saleData, saleId, isReprint) {
        const printWindow = getReceiptWindow();
        if (!printWindow) return;
        let itemsRowsHTML = '';
        (saleData.items || []).forEach(function (item) {
            const name = item.product_name || 'Item #' + item.product_id;
            itemsRowsHTML += '<tr><td style="padding:2px 0;vertical-align:top">' + name + '<br/><span style="font-size:10px;color:#555">' + item.quantity + ' x KES ' + item.unit_price.toFixed(2) + '</span></td><td style="text-align:right;padding:2px 0;vertical-align:top">' + item.subtotal.toFixed(2) + '</td></tr>';
        });
        const now = new Date().toLocaleString('en-KE', { dateStyle: 'short', timeStyle: 'short' });
        const reprintHeader = isReprint ? '<div style="text-align:center;font-weight:bold;border:1px dashed #000;padding:4px;margin-bottom:6px;font-size:12px">*** REPRINTED RECEIPT ***</div>' : '';
        const companyName = localStorage.getItem('company_name') || '🕷️ SPIDE POS';
        const companyPhone = localStorage.getItem('company_phone') || '';
        const companyEmail = localStorage.getItem('company_email') || '';
        const companyAddress = localStorage.getItem('company_address') || '';
        const companyLogo = localStorage.getItem('company_logo') || '';
        const companyFooter = localStorage.getItem('company_receipt_footer') || 'Thank you for your business!';
        const depositUsed = saleData.deposit_amount || 0;
        const creditUsed = saleData.credit_amount || 0;

        let receiptHTML = '<!DOCTYPE html><html><head><title>Receipt #' + saleId + '</title>';
        receiptHTML += '<style>@page{margin:0}body{font-family:"Courier New",monospace;width:260px;margin:10px auto;font-size:11px;color:#000;line-height:1.2}.center{text-align:center}.right{text-align:right}.dashed{border-bottom:1px dashed #000;margin:6px 0}table{width:100%;border-collapse:collapse}.bold{font-weight:bold}.shop-name{font-size:14px;font-weight:bold}</style>';
        receiptHTML += '</head><body>';
        receiptHTML += reprintHeader;
        receiptHTML += '<div class="center">';
        if (companyLogo) receiptHTML += '<img src="' + companyLogo + '" style="max-width:100px;margin:0 auto 5px">';
        receiptHTML += '<div class="shop-name">' + companyName + '</div>';
        if (companyPhone) receiptHTML += '<div style="font-size:9px">📞 ' + companyPhone + '</div>';
        if (companyEmail) receiptHTML += '<div style="font-size:9px">✉️ ' + companyEmail + '</div>';
        if (companyAddress) receiptHTML += '<div style="font-size:9px">📍 ' + companyAddress + '</div>';
        receiptHTML += '<p style="margin:4px 0">Receipt #: <strong>' + saleId + '</strong></p>';
        receiptHTML += '<p style="margin:2px 0;font-size:10px">Date: ' + now + '</p>';
        receiptHTML += '</div>';
        receiptHTML += '<div class="dashed"></div>';
        receiptHTML += '<table><thead><tr style="text-align:left;border-bottom:1px solid #000"><th>Item</th><th style="text-align:right">Subtotal</th></tr></thead><tbody>' + itemsRowsHTML + '</tbody></table>';
        receiptHTML += '<div class="dashed"></div>';
        receiptHTML += '<table><tr class="bold"><td>TOTAL:</td><td class="right">KES ' + (saleData.total_amount || 0).toFixed(2) + '</td></tr>';
        if (saleData.cash_amount > 0) receiptHTML += '<tr><td>Cash:</td><td class="right">KES ' + saleData.cash_amount.toFixed(2) + '</td></tr>';
        if (saleData.mpesa_amount > 0) receiptHTML += '<tr><td>M-Pesa:</td><td class="right">KES ' + saleData.mpesa_amount.toFixed(2) + '</td></tr>';
        if (saleData.change_given > 0) receiptHTML += '<tr><td>Change:</td><td class="right">KES ' + saleData.change_given.toFixed(2) + '</td></tr>';
        if (depositUsed > 0) receiptHTML += '<tr><td>Deposit Used:</td><td class="right">KES ' + depositUsed.toFixed(2) + '</td></tr>';
        if (creditUsed > 0) receiptHTML += '<tr><td>Credit:</td><td class="right">KES ' + creditUsed.toFixed(2) + '</td></tr>';
        receiptHTML += '</table>';
        receiptHTML += '<div class="dashed"></div>';
        receiptHTML += '<div class="center" style="margin-top:8px;font-size:10px"><p style="margin:2px 0">' + companyFooter + '</p><p style="margin:2px 0">Goods once sold are not returnable.</p></div>';
        receiptHTML += '<script>window.onload=function(){window.print();setTimeout(function(){window.close()},500)};<\/script>';
        receiptHTML += '</body></html>';
        printWindow.document.open();
        printWindow.document.write(receiptHTML);
        printWindow.document.close();
        printWindow.focus();
    };

    // =========================================================
    // TRANSFERS
    // =========================================================
    let transferItems = [];   // temp state for the create form
    let currentTransferId = null;
    let currentTransferData = null;

    window.loadTransferShops = async function () {
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/shops', {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!response.ok) return;
            const data = await response.json();
            const shops = Array.isArray(data) ? data : [];
            const select = document.getElementById('transfer-to-shop');
            if (!select) return;
            const currentValue = select.value;

            // Exclude the source shop
            const fromShopWrapper = document.getElementById('transfer-from-shop-wrapper');
            const fromShopSelect = document.getElementById('transfer-from-shop');
            let excludeId = window.getCurrentShopId();
            if (fromShopWrapper && !fromShopWrapper.classList.contains('hidden') && fromShopSelect && fromShopSelect.value) {
                excludeId = parseInt(fromShopSelect.value) || 0;
            }

            select.innerHTML = '<option value="">Select Shop</option>';
            shops.forEach(s => {
                if (s.id === excludeId) return;
                const opt = document.createElement('option');
                opt.value = s.id;
                opt.textContent = s.name;
                select.appendChild(opt);
            });
            if (currentValue && currentValue !== String(excludeId)) select.value = currentValue;
        } catch (e) {
            console.error('loadTransferShops error:', e);
        }
    };

    window.handleTransferSearch = function (query) {
        const dropdown = document.getElementById('transfer-search-results');
        if (!dropdown) return;
        const trimmed = (query || '').trim();
        if (trimmed.length < 2) { dropdown.classList.add('hidden'); return; }

        const token = window.getCookie('spide_token');
        const shopId = window.getCurrentShopId();
        const url = '/api/products/search?q=' + encodeURIComponent(trimmed) + '&shop_id=' + shopId;

        fetch(url, { headers: token ? { 'Authorization': 'Bearer ' + token } : {} })
            .then(r => r.ok ? r.json() : [])
            .then(products => {
                const list = Array.isArray(products) ? products : [];
                if (list.length === 0) {
                    dropdown.innerHTML = '<div class="p-3 text-center text-sm text-gray-500">No products found</div>';
                    dropdown.classList.remove('hidden');
                    return;
                }
                let html = '';
                list.slice(0, 15).forEach((p, i) => {
                    html += `<div onclick="selectTransferProduct(${i})" data-idx="${i}" class="transfer-search-item p-2 hover:bg-blue-50 cursor-pointer border-b last:border-b-0">
                        <div class="font-semibold text-sm text-gray-900">${p.name}</div>
                        <div class="text-xs text-gray-500">${p.barcode || 'no barcode'} • Stock: ${p.stock_quantity || 0} • Cost: KES ${(p.cost_price || 0).toFixed(2)}</div>
                    </div>`;
                });
                dropdown.innerHTML = html;
                dropdown.classList.remove('hidden');
                window.__transferSearchResults = list;
            })
            .catch(e => { console.error('handleTransferSearch error:', e); });
    };

    window.handleTransferSearchKeydown = function (e) {
        if (e.key === 'Escape') {
            const dropdown = document.getElementById('transfer-search-results');
            if (dropdown) dropdown.classList.add('hidden');
        }
    };

    window.selectTransferProduct = function (index) {
        const list = window.__transferSearchResults || [];
        const p = list[index];
        if (!p) return;

        const searchInput = document.getElementById('transfer-product-search');
        if (searchInput) searchInput.value = '';
        const dropdown = document.getElementById('transfer-search-results');
        if (dropdown) dropdown.classList.add('hidden');

        const qtyInput = document.getElementById('transfer-qty');
        const qty = parseInt(qtyInput?.value) || 1;

        window.addTransferItemByProduct(p, qty);
    };

    window.addTransferItemByProduct = function (product, qty) {
        qty = parseInt(qty) || 1;
        if (qty <= 0) { alert('Quantity must be positive'); return; }

        const available = parseInt(product.stock_quantity) || 0;
        if (available <= 0) {
            alert('This product has no stock in the source shop.');
            return;
        }

        const existingIdx = transferItems.findIndex(it => it.product_id === product.id);
        if (existingIdx >= 0) {
            transferItems[existingIdx].quantity += qty;
        } else {
            transferItems.push({
                product_id: product.id,
                product_name: product.name,
                barcode: product.barcode || '',
                quantity: qty,
                cost_price: parseFloat(product.cost_price) || 0,
                stock_available: parseInt(product.stock_quantity) || 0
            });
        }
        window.renderTransferItems();
        const qtyInput = document.getElementById('transfer-qty');
        if (qtyInput) qtyInput.value = 1;
    };

    window.addTransferItem = function () {
        const qtyInput = document.getElementById('transfer-qty');
        const qty = parseInt(qtyInput?.value) || 1;
        const searchInput = document.getElementById('transfer-product-search');
        const term = (searchInput?.value || '').trim();
        if (!term) { alert('Search for a product first'); return; }

        // If the search dropdown is showing, use the currently typed term to trigger a lookup
        const token = window.getCookie('spide_token');
        const shopId = window.getCurrentShopId();
        fetch('/api/products/search?q=' + encodeURIComponent(term) + '&shop_id=' + shopId, {
            headers: token ? { 'Authorization': 'Bearer ' + token } : {}
        })
            .then(r => r.ok ? r.json() : [])
            .then(list => {
                if (!Array.isArray(list) || list.length === 0) {
                    alert('No products match "' + term + '"');
                    return;
                }
                // Use the first match
                window.addTransferItemByProduct(list[0], qty);
                if (searchInput) searchInput.value = '';
                const dropdown = document.getElementById('transfer-search-results');
                if (dropdown) dropdown.classList.add('hidden');
            })
            .catch(e => { console.error('addTransferItem error:', e); });
    };

    window.removeTransferItem = function (index) {
        transferItems.splice(index, 1);
        window.renderTransferItems();
    };

    window.renderTransferItems = function () {
        const container = document.getElementById('transfer-items-list');
        if (!container) return;
        if (transferItems.length === 0) {
            container.innerHTML = '<p class="text-xs text-gray-400 text-center py-2">No items added</p>';
            return;
        }
        let html = '';
        transferItems.forEach((item, i) => {
            const subtotal = item.quantity * item.cost_price;
            html += `<div class="flex justify-between items-center bg-white p-2 rounded-lg border border-gray-200 gap-2">
            <div class="flex-1 min-w-0">
                <div class="text-sm font-semibold text-gray-800 truncate">${item.product_name}</div>
                <div class="text-xs text-gray-500">
                    Available: ${item.stock_available || 0} • KES ${item.cost_price.toFixed(2)} each
                </div>
            </div>
            <div class="flex items-center gap-1 flex-shrink-0">
                <button type="button" onclick="adjustTransferItemQty(${i}, -1)"
                    class="w-7 h-7 rounded-lg bg-gray-100 hover:bg-gray-200 text-gray-700 font-bold text-sm">−</button>
                <input type="number" min="1" value="${item.quantity}"
                    oninput="setTransferItemQty(${i}, this.value)"
                    class="w-16 px-2 py-1 border rounded-lg text-sm text-center focus:ring-2 focus:ring-blue-500 focus:outline-none">
                <button type="button" onclick="adjustTransferItemQty(${i}, 1)"
                    class="w-7 h-7 rounded-lg bg-gray-100 hover:bg-gray-200 text-gray-700 font-bold text-sm">+</button>
            </div>
            <div class="w-20 text-right text-sm font-semibold text-gray-800 flex-shrink-0">
                KES ${subtotal.toFixed(2)}
            </div>
            <button type="button" onclick="removeTransferItem(${i})"
                class="text-red-500 hover:text-red-700 text-sm flex-shrink-0">✕</button>
        </div>`;
        });
        container.innerHTML = html;
    };

    window.adjustTransferItemQty = function (index, delta) {
        if (!transferItems[index]) return;
        const newQty = transferItems[index].quantity + delta;
        setTransferItemQty(index, newQty);
    };

    window.setTransferItemQty = function (index, value) {
        if (!transferItems[index]) return;
        const qty = parseInt(value) || 0;
        const available = transferItems[index].stock_available || 0;

        if (qty <= 0) {
            // Remove item if quantity goes to 0 or below
            transferItems.splice(index, 1);
            window.renderTransferItems();
            return;
        }
        if (available > 0 && qty > available) {
            transferItems[index].quantity = available;
            window.renderTransferItems();
            if (typeof window.showNotification === 'function') {
                window.showNotification('Only ' + available + ' available in stock', 'warning');
            }
            return;
        }
        transferItems[index].quantity = qty;
        // Re-render only the subtotal — but full re-render is simpler and fast enough
        window.renderTransferItems();
    };

    window.submitTransfer = async function (event) {
        event.preventDefault();

        // 1. Grab all the DOM inputs
        const alertBox = document.getElementById('transfer-alert');
        const submitBtn = document.getElementById('transfer-submit-btn');
        const toShopEl = document.getElementById('transfer-to-shop');
        const dateEl = document.getElementById('transfer-date');
        const notesEl = document.getElementById('transfer-notes');

        // 2. setAlert helper
        const setAlert = (msg, type) => {
            if (!alertBox) return;
            alertBox.className = 'p-2 rounded-lg text-sm font-medium ' +
                (type === 'success' ? 'bg-green-100 text-green-800 border border-green-200'
                    : 'bg-red-100 text-red-800 border border-red-200');
            alertBox.textContent = msg;
            alertBox.classList.remove('hidden');
        };

        // 3. Resolve source shop FIRST (before any validation that uses it)
        const fromShopWrapper = document.getElementById('transfer-from-shop-wrapper');
        const fromShopSelect = document.getElementById('transfer-from-shop');
        const role = typeof window.getCurrentUserRole === 'function' ? window.getCurrentUserRole() : 'cashier';
        const isDirector = (role === 'director' || role === 'admin');

        let fromShopId;
        if (isDirector && fromShopWrapper && !fromShopWrapper.classList.contains('hidden')) {
            fromShopId = parseInt(fromShopSelect?.value) || 0;
            if (!fromShopId) { setAlert('Select a source shop', 'error'); return; }
        } else {
            fromShopId = window.getCurrentShopId();
        }

        // 4. Now resolve destination + other inputs
        const toShopId = parseInt(toShopEl?.value) || 0;

        // 5. Validate
        if (!toShopId) { setAlert('Select a destination shop', 'error'); return; }
        if (toShopId === fromShopId) { setAlert('Source and destination must differ', 'error'); return; }
        if (transferItems.length === 0) { setAlert('Add at least one item', 'error'); return; }

        // 6. Build payload
        const payload = {
            from_shop_id: fromShopId,
            to_shop_id: toShopId,
            transfer_date: dateEl?.value || new Date().toISOString().split('T')[0],
            notes: notesEl?.value || '',
            status: 'completed',
            items: transferItems.map(it => ({
                product_id: it.product_id,
                quantity: it.quantity,
                cost_price: it.cost_price
            }))
        };

        // 7. Submit
        if (submitBtn) { submitBtn.disabled = true; submitBtn.textContent = 'Creating...'; }
        alertBox?.classList.add('hidden');

        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/transfers/create', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...(token ? { 'Authorization': 'Bearer ' + token } : {})
                },
                body: JSON.stringify(payload)
            });
            const rawText = await response.text();
            let result = {};
            try { result = rawText ? JSON.parse(rawText) : {}; } catch (e) { }

            if (response.ok) {
                setAlert('✅ Transfer created!', 'success');

                // ✅ Print receipt before closing the modal
                try {
                    const createdId = result.id || result.transfer?.id;
                    if (createdId && typeof window.printTransferReceiptById === 'function') {
                        await window.printTransferReceiptById(createdId);
                    } else if (createdId && typeof window.printTransferReceipt === 'function') {
                        // Fallback: build the data from the current form if the by-id version isn't available
                        await window.printTransferReceipt(createdId);
                    }
                } catch (e) {
                    console.error('Auto-print failed:', e);
                }

                if (typeof window.resetTransferItems === 'function') window.resetTransferItems();
                else { transferItems = []; window.renderTransferItems(); }
                if (toShopEl) toShopEl.value = '';
                if (notesEl) notesEl.value = '';
                window.__transferSearchResults = [];
                setTimeout(() => {
                    if (typeof window.closeTransferModal === 'function') window.closeTransferModal();
                }, 500);

            } else {
                setAlert('Error: ' + (result.error || rawText || 'Unknown error'), 'error');
            }
        } catch (e) {
            console.error('submitTransfer error:', e);
            setAlert('Network error: ' + e.message, 'error');
        } finally {
            if (submitBtn) { submitBtn.disabled = false; submitBtn.textContent = '📦 Create Transfer'; }
        }
    };

    window.loadTransferHistory = async function () {
        const tbody = document.getElementById('transfer-history-body');
        if (!tbody) return;
        tbody.innerHTML = '<tr><td colspan="8" class="p-4 text-center text-gray-500">Loading...</td></tr>';
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/transfers', {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!response.ok) throw new Error('Failed to load transfers: ' + response.status);
            const data = await response.json();
            const list = Array.isArray(data) ? data : [];
            if (list.length === 0) {
                tbody.innerHTML = '<tr><td colspan="8" class="p-4 text-center text-gray-500">No transfers yet</td></tr>';
                return;
            }
            let html = '';
            list.forEach(t => {
                const statusColor = t.status === 'completed' ? 'bg-green-100 text-green-800'
                    : t.status === 'pending' ? 'bg-amber-100 text-amber-800'
                        : 'bg-red-100 text-red-800';
                html += `<tr class="border-b hover:bg-blue-50/50 transition">
                    <td class="p-2 font-mono text-xs">${t.transfer_number}</td>
                    <td class="p-2">${t.from_shop_name || '#' + t.from_shop_id}</td>
                    <td class="p-2">${t.to_shop_name || '#' + t.to_shop_id}</td>
                    <td class="p-2 text-center">${t.total_items || 0}</td>
                    <td class="p-2 text-right font-mono">KES ${(t.total_cost || 0).toFixed(2)}</td>
                    <td class="p-2 text-xs text-gray-500">${t.transfer_date || ''}</td>
                    <td class="p-2 text-center"><span class="px-2 py-0.5 rounded-full text-xs ${statusColor}">${t.status}</span></td>
                    <td class="p-2 text-center">
                        <button onclick="openTransferDetailModal(${t.id})" class="text-blue-500 hover:text-blue-700 text-xs font-bold">👁️ View</button>
                    </td>
                </tr>`;
            });
            tbody.innerHTML = html;
        } catch (e) {
            console.error('loadTransferHistory error:', e);
            tbody.innerHTML = '<tr><td colspan="8" class="p-4 text-center text-red-500">Error: ' + e.message + '</td></tr>';
        }
    };

    window.openTransferDetailModal = async function (id) {
        const modal = document.getElementById('transfer-detail-modal');
        if (!modal) return;
        modal.classList.remove('hidden');
        modal.classList.add('flex');

        // Reset UI
        const setText = (elId, text) => { const el = document.getElementById(elId); if (el) el.textContent = text; };
        setText('transfer-detail-number', '#TRF-' + id);
        setText('td-from-shop', '...');
        setText('td-to-shop', '...');
        setText('td-date', '...');
        setText('td-status', '...');
        setText('td-total-items', '0');
        setText('td-total-cost', 'KES 0.00');
        const tbody = document.getElementById('transfer-detail-items');
        if (tbody) tbody.innerHTML = '<tr><td colspan="5" class="p-4 text-center text-gray-500">Loading...</td></tr>';

        currentTransferId = id;
        currentTransferData = t;

        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/transfers/detail?id=' + id, {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!response.ok) throw new Error('Failed to load transfer');
            const t = await response.json();
            setText('transfer-detail-number', t.transfer_number || ('#TRF-' + id));
            setText('td-from-shop', t.from_shop_name || '#' + t.from_shop_id);
            setText('td-to-shop', t.to_shop_name || '#' + t.to_shop_id);
            setText('td-date', t.transfer_date || '');
            setText('td-status', t.status || '');
            setText('td-total-items', String(t.total_items || 0));
            setText('td-total-cost', 'KES ' + (t.total_cost || 0).toFixed(2));

            const items = t.items || [];
            if (tbody) {
                if (items.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="5" class="p-4 text-center text-gray-500">No items</td></tr>';
                } else {
                    let html = '';
                    items.forEach(it => {
                        html += `<tr class="border-b">
                            <td class="p-2 font-medium text-gray-800">${it.product_name}</td>
                            <td class="p-2 font-mono text-xs text-gray-500">${it.barcode || 'N/A'}</td>
                            <td class="p-2 text-center font-bold">${it.quantity}</td>
                            <td class="p-2 text-right font-mono">KES ${(it.cost_price || 0).toFixed(2)}</td>
                            <td class="p-2 text-right font-mono">KES ${(it.subtotal || 0).toFixed(2)}</td>
                        </tr>`;
                    });
                    tbody.innerHTML = html;
                }
            }
        } catch (e) {
            console.error('openTransferDetailModal error:', e);
            if (tbody) tbody.innerHTML = '<tr><td colspan="5" class="p-4 text-center text-red-500">' + e.message + '</td></tr>';
        }
    };

    window.closeTransferHistoryModal = function () {
        const modal = document.getElementById('transfer-history-modal');
        if (modal) modal.classList.add('hidden');
    };

    window.showFromShopIfDirector = function () {
        const wrapper = document.getElementById('transfer-from-shop-wrapper');
        if (!wrapper) return;

        const role = typeof window.getCurrentUserRole === 'function' ? window.getCurrentUserRole() : 'cashier';
        if (role === 'director' || role === 'admin') {
            wrapper.classList.remove('hidden');
            window.populateTransferFromShop();
        } else {
            wrapper.classList.add('hidden');
        }
    };

    window.populateTransferFromShop = async function () {
        const select = document.getElementById('transfer-from-shop');
        if (!select) return;
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/shops', {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!response.ok) return;
            const data = await response.json();
            const shops = Array.isArray(data) ? data : [];
            const currentValue = select.value;
            select.innerHTML = '<option value="">Select Source Shop</option>';
            shops.forEach(s => {
                const opt = document.createElement('option');
                opt.value = s.id;
                opt.textContent = s.name;
                select.appendChild(opt);
            });
            // Default to current shop
            if (!currentValue) {
                const currentShopId = window.getCurrentShopId();
                if (currentShopId) select.value = String(currentShopId);
            } else {
                select.value = currentValue;
            }
        } catch (e) {
            console.error('populateTransferFromShop error:', e);
        }
    };

    window.printTransferReceipt = function (id) {
        const t = currentTransferData;
        if (!t) {
            alert('No transfer loaded to print.');
            return;
        }

        const printWindow = window.open('', '_blank', 'width=350,height=600');
        if (!printWindow) return;

        const companyName = localStorage.getItem('company_name') || '🕷️ SPIDE POS';
        const companyPhone = localStorage.getItem('company_phone') || '';
        const companyAddress = localStorage.getItem('company_address') || '';
        const now = new Date().toLocaleString('en-KE', { dateStyle: 'short', timeStyle: 'short' });

        let itemRows = '';
        (t.items || []).forEach(it => {
            itemRows += `<tr>
            <td style="padding:2px 0;vertical-align:top">${it.product_name}<br/>
                <span style="font-size:10px;color:#555">${it.quantity} × KES ${(it.cost_price || 0).toFixed(2)}</span>
            </td>
            <td style="text-align:right;padding:2px 0;vertical-align:top">
                ${(it.subtotal || 0).toFixed(2)}
            </td>
        </tr>`;
        });

        let html = '<!DOCTYPE html><html><head><title>Transfer ' + (t.transfer_number || id) + '</title>';
        html += '<style>@page{margin:0}body{font-family:"Courier New",monospace;width:260px;margin:10px auto;font-size:11px;color:#000;line-height:1.2}.center{text-align:center}.right{text-align:right}.dashed{border-bottom:1px dashed #000;margin:6px 0}table{width:100%;border-collapse:collapse}.bold{font-weight:bold}.shop-name{font-size:14px;font-weight:bold}</style>';
        html += '</head><body>';
        html += '<div class="center">';
        html += '<div class="shop-name">' + companyName + '</div>';
        if (companyPhone) html += '<div style="font-size:9px">📞 ' + companyPhone + '</div>';
        if (companyAddress) html += '<div style="font-size:9px">📍 ' + companyAddress + '</div>';
        html += '<div style="font-weight:bold;margin-top:6px;font-size:12px">STOCK TRANSFER</div>';
        html += '<p style="margin:4px 0">Transfer #: <strong>' + (t.transfer_number || id) + '</strong></p>';
        html += '<p style="margin:2px 0;font-size:10px">Date: ' + (t.transfer_date || now) + '</p>';
        html += '<p style="margin:2px 0;font-size:10px">Printed: ' + now + '</p>';
        html += '</div>';
        html += '<div class="dashed"></div>';
        html += '<div style="font-size:10px">';
        html += '<div><strong>From:</strong> ' + (t.from_shop_name || '#' + t.from_shop_id) + '</div>';
        html += '<div><strong>To:</strong> ' + (t.to_shop_name || '#' + t.to_shop_id) + '</div>';
        if (t.created_by) html += '<div><strong>By:</strong> ' + t.created_by + '</div>';
        if (t.notes) html += '<div><strong>Notes:</strong> ' + t.notes + '</div>';
        html += '</div>';
        html += '<div class="dashed"></div>';
        html += '<table><thead><tr style="text-align:left;border-bottom:1px solid #000"><th>Item</th><th style="text-align:right">Subtotal</th></tr></thead><tbody>' + itemRows + '</tbody></table>';
        html += '<div class="dashed"></div>';
        html += '<table>';
        html += '<tr class="bold"><td>TOTAL ITEMS:</td><td class="right">' + (t.total_items || 0) + '</td></tr>';
        html += '<tr class="bold"><td>TOTAL COST:</td><td class="right">KES ' + (t.total_cost || 0).toFixed(2) + '</td></tr>';
        html += '</table>';
        html += '<div class="dashed"></div>';
        html += '<div class="center" style="margin-top:8px;font-size:10px">';
        html += '<p style="margin:2px 0">Transfer Receipt</p>';
        html += '</div>';
        html += '<script>window.onload=function(){window.print();setTimeout(function(){window.close()},500)};<\/script>';
        html += '</body></html>';

        printWindow.document.write(html);
        printWindow.document.close();
    };

    window.printTransferReceiptById = async function (id) {
        if (!id) {
            alert('No transfer ID to print.');
            return;
        }
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/transfers/detail?id=' + id, {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!response.ok) {
                const errText = await response.text();
                throw new Error(errText || 'Failed to load transfer for printing');
            }
            const data = await response.json();

            // Store so printTransferReceipt can access it
            currentTransferId = id;
            currentTransferData = data;

            // Now call the existing print function
            if (typeof window.printTransferReceipt === 'function') {
                window.printTransferReceipt(id);
            }
        } catch (e) {
            console.error('printTransferReceiptById error:', e);
            // Don't alert — auto-print failing shouldn't block the user
        }
    };

    window.removeTransferItem = window.removeTransferItem || function (index) {
        transferItems.splice(index, 1);
        window.renderTransferItems();
    };

    // =========================================================
    // PRODUCTS / CATALOG
    // =========================================================
    window.__allProducts = window.__allProducts || [];
    window.__filteredProducts = window.__filteredProducts || [];

    window.loadProducts = async function () {
        const tbody = document.getElementById('product-catalog-rows');
        if (!tbody) return;
        tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-gray-500">Loading products...</td></tr>';
        try {
            const shopFilterEl = document.getElementById('catalog-shop-filter');
            const shopFilterWrapper = document.getElementById('catalog-shop-filter-wrapper');
            let shopId = window.getCurrentShopId();
            if (shopFilterEl && shopFilterWrapper && !shopFilterWrapper.classList.contains('hidden')) {
                shopId = parseInt(shopFilterEl.value) || 0;
            }
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/products?shop_id=' + shopId, {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!response.ok) throw new Error('Failed: ' + response.status);
            const products = await response.json();
            window.__allProducts = products || [];
            window.__filteredProducts = [...window.__allProducts];
            window.updateCatalogCounts();
            // ✅ Refresh category filter options
            if (typeof window.populateCatalogCategoryFilter === 'function') {
                window.populateCatalogCategoryFilter();
            }

            if (window.__allProducts.length === 0) {
                tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-gray-500">No products found.</td></tr>';
                const c = document.getElementById('catalog-total-count');
                if (c) c.textContent = '0 products';
                return;
            }
            window.renderProductCatalog(window.__allProducts);
        } catch (err) {
            tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-red-500">Error: ' + err.message + '</td></tr>';
        }
    };

    window.renderProductCatalog = function (products) {
        const tbody = document.getElementById('product-catalog-rows');
        if (!tbody) return;
        if (!products || products.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-gray-500">No products</td></tr>';
            return;
        }
        let html = '';
        products.forEach(p => {
            const barcodeDisplay = p.barcode || 'N/A';
            const stock = p.stock_quantity || 0;
            const isLow = stock <= p.reorder_level && stock > 0;
            const isOut = stock <= 0;
            const role = typeof window.getCurrentUserRole === 'function' ? window.getCurrentUserRole() : 'cashier';
            const canEdit = (role === 'director' || role === 'admin' || role === 'manager');

            let stockBadge, stockClass;
            if (isOut) { stockBadge = '<span class="bg-red-100 text-red-800 text-xs font-bold px-2 py-0.5 rounded-full">Out</span>'; stockClass = 'text-red-600 font-bold'; }
            else if (isLow) { stockBadge = '<span class="bg-amber-100 text-amber-800 text-xs font-bold px-2 py-0.5 rounded-full">' + stock + ' (Low)</span>'; stockClass = 'text-amber-600 font-bold'; }
            else { stockBadge = '<span class="bg-emerald-100 text-emerald-800 text-xs font-bold px-2 py-0.5 rounded-full">' + stock + '</span>'; stockClass = 'text-emerald-600'; }
            html += `<tr class="border-b hover:bg-purple-50/50 transition" data-product-id="${p.id}">
                <td class="p-3 font-mono text-xs text-gray-500 font-semibold">${barcodeDisplay}</td>
                <td class="p-3 font-semibold text-gray-900">${p.name}</td>
                <td class="p-3 text-gray-500 text-xs">${p.category || 'General'}</td>
                <td class="p-3 text-right font-mono font-bold text-purple-900">KES ${(p.retail_price || 0).toFixed(2)}</td>
                <td class="p-3 text-right font-mono text-gray-700">KES ${(p.wholesale_price || 0).toFixed(2)}</td>
                <td class="p-3 text-center ${stockClass}">${stockBadge}</td>
                <td class="p-3 text-center">
                    ${canEdit ? `
                        <button onclick='editProduct(${JSON.stringify(p).replace(/'/g, "&#39;")})' 
                            class="bg-blue-100 hover:bg-blue-200 text-blue-800 text-xs font-bold px-2.5 py-1 rounded-lg transition">✏️ Edit</button>
                        <button onclick='quickUpdateStock(${p.id}, "${(p.name || '').replace(/"/g, '\\"')}", ${stock})' 
                            class="bg-emerald-100 hover:bg-emerald-200 text-emerald-800 text-xs font-bold px-2.5 py-1 rounded-lg transition ml-1">📦 Stock</button>
                    ` : '<span class="text-xs text-gray-400">—</span>'}
                </td>
            </tr>`;
        });
        tbody.innerHTML = html;
    };

    window.populateCatalogShopFilter = async function () {
        const select = document.getElementById('catalog-shop-filter');
        if (!select) return;
        const currentValue = select.value;
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/shops', {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!response.ok) return;
            const shops = await response.json();
            select.innerHTML = '<option value="0">All Shops</option>';
            (Array.isArray(shops) ? shops : []).forEach(s => {
                const opt = document.createElement('option');
                opt.value = s.id;
                opt.textContent = s.name;
                select.appendChild(opt);
            });
            if (currentValue) select.value = currentValue;
        } catch (e) {
            console.error('populateCatalogShopFilter error:', e);
        }
    };

    window.populateCatalogCategoryFilter = function () {
        const select = document.getElementById('catalog-category-filter');
        if (!select) return;
        const currentValue = select.value;

        // Extract unique categories from loaded products
        const products = window.__allProducts || [];
        const categories = new Set();
        products.forEach(p => {
            const c = (p.category || '').trim();
            if (c) categories.add(c);
        });

        select.innerHTML = '<option value="">All Categories</option>';
        [...categories].sort().forEach(cat => {
            const opt = document.createElement('option');
            opt.value = cat;
            opt.textContent = cat;
            select.appendChild(opt);
        });

        // Restore previous selection if it still exists
        if (currentValue) {
            const stillExists = [...select.options].some(o => o.value === currentValue);
            if (stillExists) select.value = currentValue;
            else select.value = '';
        }
    };

    // =========================================================
    // QUICK STOCK UPDATE
    // =========================================================
    window.quickUpdateStock = function (productId, productName, currentStock) {
        const existing = document.getElementById('stock-update-modal');
        if (existing) existing.remove();
        const modal = document.createElement('div');
        modal.id = 'stock-update-modal';
        modal.className = 'fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4';
        modal.innerHTML = `
            <div class="bg-white rounded-2xl shadow-2xl w-full max-w-md p-6 space-y-4">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2"><span class="bg-emerald-100 p-1.5 rounded-lg">📦</span>Update Stock</h2>
                    <button onclick="closeStockUpdate()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>
                <div>
                    <p class="text-sm font-medium text-gray-700">Product: <span class="text-purple-700">${productName}</span></p>
                    <p class="text-xs text-gray-500 mt-1">Current Stock: <span class="font-bold">${currentStock}</span></p>
                </div>
                <div>
                    <label class="block text-xs font-semibold text-gray-700 uppercase mb-1">New Stock Quantity</label>
                    <input type="number" id="stock-quantity-input" value="${currentStock}" min="0" class="w-full px-3 py-2 border rounded-lg text-lg focus:ring-2 focus:ring-emerald-500 focus:outline-none">
                </div>
                <div>
                    <label class="block text-xs font-semibold text-gray-700 uppercase mb-1">Adjustment Type</label>
                    <select id="stock-adjustment-type" class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none">
                        <option value="set">Set Exact Quantity</option>
                        <option value="add">Add to Current</option>
                        <option value="subtract">Subtract from Current</option>
                    </select>
                </div>
                <div id="stock-update-alert" class="hidden p-2 rounded-lg text-sm font-medium"></div>
                <div class="flex gap-3 pt-4 border-t border-gray-100">
                    <button onclick="closeStockUpdate()" class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">Cancel</button>
                    <button onclick="saveStockUpdate(${productId})" class="flex-1 bg-emerald-600 hover:bg-emerald-700 text-white font-medium py-2 rounded-lg transition text-sm">💾 Update Stock</button>
                </div>
            </div>`;
        document.body.appendChild(modal);
        modal.classList.remove('hidden');
        setTimeout(() => { document.getElementById('stock-quantity-input')?.focus(); document.getElementById('stock-quantity-input')?.select(); }, 100);
    };

    window.closeStockUpdate = function () {
        const modal = document.getElementById('stock-update-modal');
        if (modal) { modal.classList.add('hidden'); setTimeout(() => modal.remove(), 300); }
    };

    window.saveStockUpdate = async function (productId) {
        const quantityInput = document.getElementById('stock-quantity-input');
        const adjustmentType = document.getElementById('stock-adjustment-type').value;
        const alertBox = document.getElementById('stock-update-alert');
        const submitBtn = document.querySelector('#stock-update-modal .bg-emerald-600');
        if (!quantityInput) return;
        const newQuantity = parseInt(quantityInput.value) || 0;
        if (newQuantity < 0) { window.showStockAlert('Quantity cannot be negative', 'error'); return; }
        if (submitBtn) { submitBtn.disabled = true; submitBtn.textContent = 'Saving...'; }
        alertBox?.classList.add('hidden');
        try {
            const token = window.getCookie('spide_token');
            if (!token) { window.showStockAlert('❌ Not authenticated.', 'error'); return; }
            const shopId = window.getCurrentShopId();
            const payload = {
                product_id: productId,
                quantity: newQuantity,
                adjustment_type: adjustmentType,
                shop_id: parseInt(shopId) || 1
            };
            const response = await fetch('/api/products/update-stock', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token },
                body: JSON.stringify(payload)
            });
            if (!response.ok) {
                const text = await response.text();
                window.showStockAlert('❌ Error: ' + (text || 'Unknown error'), 'error');
                return;
            }
            const responseText = await response.text();
            if (!responseText || responseText.trim() === '') { window.showStockAlert('❌ Empty response', 'error'); return; }
            let result;
            try { result = JSON.parse(responseText); } catch (e) { window.showStockAlert('❌ Invalid response', 'error'); return; }
            if (result.success) {
                window.showStockAlert('✅ Stock updated! New quantity: ' + result.new_quantity, 'success');
                setTimeout(() => {
                    window.closeStockUpdate();
                    if (typeof window.loadProducts === 'function') window.loadProducts();
                }, 1500);
            } else {
                window.showStockAlert('❌ Error: ' + (result.error || result.message || 'Unknown'), 'error');
            }
        } catch (error) {
            console.error('Stock update error:', error);
            window.showStockAlert('❌ Network error: ' + error.message, 'error');
        } finally {
            if (submitBtn) { submitBtn.disabled = false; submitBtn.textContent = '💾 Update Stock'; }
        }
    };

    window.showStockAlert = function (message, type) {
        const alertBox = document.getElementById('stock-update-alert');
        if (!alertBox) return;
        alertBox.textContent = message;
        alertBox.className = 'p-2 rounded-lg text-sm font-medium ' + (type === 'success' ? 'bg-green-100 text-green-700 border border-green-200' : type === 'error' ? 'bg-red-100 text-red-700 border border-red-200' : 'bg-amber-100 text-amber-700 border border-amber-200');
        alertBox.classList.remove('hidden');
    };

    window.filterCatalog = function () {
        const searchInput = document.getElementById('catalog-search-input');
        const categorySelect = document.getElementById('catalog-category-filter');

        const searchTerm = searchInput ? searchInput.value.toLowerCase().trim() : '';
        const categoryTerm = categorySelect ? categorySelect.value : '';

        if (window.__allProducts.length === 0) {
            window.loadProducts().then(() => window.filterCatalog());
            return;
        }

        window.__filteredProducts = window.__allProducts.filter(product => {
            // Category filter (exact match)
            if (categoryTerm) {
                const cat = (product.category || '').toLowerCase();
                if (cat !== categoryTerm.toLowerCase()) return false;
            }
            // Search filter (fuzzy)
            if (searchTerm) {
                const nameMatch = product.name?.toLowerCase().includes(searchTerm) || false;
                const barcodeMatch = (product.barcode || '').toLowerCase().includes(searchTerm);
                const categoryMatch = product.category?.toLowerCase().includes(searchTerm) || false;
                if (!nameMatch && !barcodeMatch && !categoryMatch) return false;
            }
            return true;
        });

        // Update count display
        const countEl = document.getElementById('catalog-search-count');
        if (countEl) {
            if (searchTerm || categoryTerm) {
                countEl.textContent = `${window.__filteredProducts.length} result${window.__filteredProducts.length !== 1 ? 's' : ''}`;
            } else {
                countEl.textContent = '';
            }
        }

        window.renderProductCatalog(window.__filteredProducts);
        window.updateCatalogCounts();
    };

    window.updateCatalogCounts = function () {
        const el = document.getElementById('catalog-total-count');
        if (el) el.textContent = `Showing ${window.__filteredProducts.length} of ${window.__allProducts.length} products`;
    };

    window.clearCatalogSearch = function () {
        const searchInput = document.getElementById('catalog-search-input');
        if (searchInput) { searchInput.value = ''; window.filterCatalog(); searchInput.focus(); }
    };

    window.editProduct = function (product) {
        if (typeof window.closeProductCatalogModal === 'function') window.closeProductCatalogModal();
        setTimeout(function () {
            const barcode = product.barcode || '';
            const setVal = (id, v) => { const el = document.getElementById(id); if (el) el.value = v; };
            setVal('prodSku', barcode);
            setVal('prodName', product.name || '');
            setVal('prodCategory', product.category || '');
            setVal('prodBuyingPrice', product.cost_price || 0);
            setVal('prodSellingPrice', product.retail_price || 0);
            setVal('prodWholesalePrice', product.wholesale_price || 0);
            setVal('prodWholesaleQty', product.wholesale_min_qty || 0);
            setVal('prodStockQty', product.stock_quantity || 0);
            setVal('prodReorderLevel', product.reorder_level || 5);

            const form = document.getElementById('addProductForm');
            if (form) form.dataset.editId = product.id;

            const h3 = document.querySelector('#addProductModal h3');
            if (h3) h3.innerHTML = '✏️ Edit Product';
            const btn = document.getElementById('saveProductBtn');
            if (btn) btn.innerText = '💾 Update Product';

            const sku = document.getElementById('prodSku');
            if (sku && !barcode) sku.placeholder = 'No barcode (leave blank to keep)';

            // ✅ Hide the Initial Stock field — stock is managed via the 📦 Stock button
            const stockEl = document.getElementById('prodStockQty');
            if (stockEl) {
                const wrapper = stockEl.closest('div');
                if (wrapper) wrapper.style.display = 'none';
            }

            if (typeof window.openAddProductModal === 'function') window.openAddProductModal();
        }, 300);
    };

    // =========================================================
    // Z-REPORT
    // =========================================================
    window.loadZReportShops = async function () {
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/shops', { headers: token ? { 'Authorization': 'Bearer ' + token } : {} });
            if (!response.ok) return;
            const shopList = await response.json();
            const select = document.getElementById('zreport-shop-filter');
            if (!select) return;
            const currentValue = select.value;
            select.innerHTML = '<option value="0">All Shops</option>';
            shopList.forEach(b => {
                const opt = document.createElement('option');
                opt.value = b.id;
                opt.textContent = b.name;
                select.appendChild(opt);
            });
            if (currentValue) select.value = currentValue;
        } catch (e) { console.error('Error loading shops:', e); }
    };

    window.loadZReport = async function () {
        const dateEl = document.getElementById('zreport-date');
        const date = dateEl ? dateEl.value : new Date().toISOString().split('T')[0];
        const shopId = document.getElementById('zreport-shop-filter')?.value || 0;
        if (!date) { alert('Please select a date'); return; }
        const container = document.getElementById('zreport-content');
        if (!container) return;
        container.innerHTML = '<div class="text-center text-gray-500 py-8">Loading report...</div>';
        try {
            const token = window.getCookie('spide_token');
            const url = '/api/sales/z-report?date=' + date + '&shop_id=' + shopId;
            const response = await fetch(url, { headers: token ? { 'Authorization': 'Bearer ' + token } : {} });
            if (!response.ok) throw new Error('Failed to load Z-Report');
            const data = await response.json();
            window.renderZReport(data);
        } catch (error) {
            container.innerHTML = '<div class="text-center text-red-500 py-8">Error: ' + error.message + '</div>';
        }
    };

    window.renderZReport = function (data) {
        const container = document.getElementById('zreport-content');
        if (!container) return;
        if (!data || !data.total_revenue) {
            container.innerHTML = '<div class="text-center py-12"><span class="text-4xl mb-3 block">📭</span><p class="text-sm text-gray-500">No sales data for this date</p></div>';
            return;
        }
        let html = `
            <div class="text-center border-b border-gray-200 pb-4 mb-4">
                <h2 class="text-xl font-bold text-gray-900">📊 Daily Z-Report</h2>
                <p class="text-sm text-gray-500">${data.report_date || 'N/A'}</p>
                ${data.shop_name ? `<p class="text-xs text-gray-400">🏪 ${data.shop_name}</p>` : ''}
            </div>
            <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4">
                <div class="bg-purple-50 p-4 rounded-xl text-center"><p class="text-xs text-purple-600 uppercase">Revenue</p><p class="text-xl font-bold text-purple-900">KES ${(data.total_revenue || 0).toFixed(2)}</p></div>
                <div class="bg-blue-50 p-4 rounded-xl text-center"><p class="text-xs text-blue-600 uppercase">Sales</p><p class="text-xl font-bold text-blue-900">${data.total_sales_count || 0}</p></div>
                <div class="bg-amber-50 p-4 rounded-xl text-center"><p class="text-xs text-amber-600 uppercase">Expenses</p><p class="text-xl font-bold text-amber-900">KES ${(data.total_expenses || 0).toFixed(2)}</p></div>
                <div class="bg-emerald-50 p-4 rounded-xl text-center"><p class="text-xs text-emerald-600 uppercase">Net</p><p class="text-xl font-bold text-emerald-900">KES ${(data.net_profit || 0).toFixed(2)}</p></div>
            </div>
            <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
                <div class="bg-white border rounded-xl p-3 text-center"><p class="text-xs text-gray-500">💵 Cash</p><p class="font-bold">KES ${(data.total_cash || 0).toFixed(2)}</p></div>
                <div class="bg-white border rounded-xl p-3 text-center"><p class="text-xs text-gray-500">📱 M-Pesa</p><p class="font-bold">KES ${(data.total_mpesa || 0).toFixed(2)}</p></div>
                <div class="bg-white border rounded-xl p-3 text-center"><p class="text-xs text-gray-500">💰 Deposit</p><p class="font-bold">KES ${(data.total_deposit || 0).toFixed(2)}</p></div>
                <div class="bg-white border rounded-xl p-3 text-center"><p class="text-xs text-gray-500">📊 Credit</p><p class="font-bold">KES ${(data.total_credit || 0).toFixed(2)}</p></div>
            </div>`;
        container.innerHTML = html;
    };

    window.printZReport = function () {
        const content = document.getElementById('zreport-content');
        if (!content) return;
        const printWindow = window.open('', '_blank', 'width=800,height=600');
        printWindow.document.write(`<html><head><title>Z-Report</title><style>body{font-family:Arial,sans-serif;padding:20px}</style></head><body>${content.innerHTML}<script>window.onload=function(){window.print();setTimeout(function(){window.close()},500)};<\/script></body></html>`);
        printWindow.document.close();
    };

    // =========================================================
    // LOW STOCK
    // =========================================================
    window.fetchLowStockData = async function () {
        try {
            const shopId = window.getCurrentShopId();
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/reports/low-stock?threshold=5&shop_id=' + shopId, {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            const data = await response.json();
            const container = document.getElementById('low-stock-content');
            if (!container) return;
            if (!data || data.length === 0) {
                container.innerHTML = '<div class="text-center text-emerald-600 font-semibold py-6">🎉 All stock levels are healthy!</div>';
                return;
            }
            let html = `<table class="w-full text-left text-sm"><thead class="bg-gray-100 border-b"><tr><th class="p-2">Product</th><th class="p-2">Category</th><th class="p-2 text-center">Stock</th><th class="p-2 text-center">Reorder</th><th class="p-2 text-right">Est. Cost</th></tr></thead><tbody>`;
            data.forEach(item => {
                const isCritical = item.stock_quantity === 0;
                const statusClass = isCritical ? 'text-red-600 font-bold' : 'text-amber-600 font-bold';
                html += `<tr class="border-b"><td class="p-2 font-semibold">${item.product_name}</td><td class="p-2 text-gray-500">${item.category || 'General'}</td><td class="p-2 text-center ${statusClass}">${item.stock_quantity}</td><td class="p-2 text-center">${item.reorder_level}</td><td class="p-2 text-right font-mono">KES ${(item.restock_cost || 0).toFixed(2)}</td></tr>`;
            });
            html += '</tbody></table>';
            const totalRestock = data.reduce((sum, item) => sum + (item.restock_cost || 0), 0);
            html += `<div class="bg-amber-50 border-t border-amber-200 p-3 mt-2"><div class="flex justify-between items-center"><span class="font-bold text-amber-900">Total Restock:</span><span class="font-bold text-amber-900">KES ${totalRestock.toFixed(2)}</span></div><div class="text-xs text-amber-700 mt-1">${data.length} items need reordering</div></div>`;
            container.innerHTML = html;
        } catch (err) {
            const container = document.getElementById('low-stock-content');
            if (container) container.innerHTML = '<p class="text-center text-red-500 py-6">Error loading stock data</p>';
        }
    };

    // =========================================================
    // INVENTORY VALUATION
    // =========================================================
    window.loadShopsForValuation = async function () {
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/shops', { headers: token ? { 'Authorization': 'Bearer ' + token } : {} });
            if (!response.ok) return;
            const shopList = await response.json();
            const select = document.getElementById('valuation-branch-select');
            if (!select) return;
            select.innerHTML = '<option value="0">All Shops</option>';
            shopList.forEach(b => {
                const opt = document.createElement('option');
                opt.value = b.id;
                opt.textContent = b.name;
                select.appendChild(opt);
            });
        } catch (e) { console.error('Error loading shops:', e); }
    };

    window.loadInventoryValuation = async function () {
        const branchSelect = document.getElementById('valuation-branch-select');
        const branchId = branchSelect ? branchSelect.value : 0;
        const shopId = branchId || window.getCurrentShopId();
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/reports/inventory-valuation?branch_id=' + shopId, {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!response.ok) throw new Error('Failed to load valuation');
            const data = await response.json();
            window.renderValuation(data);
        } catch (error) {
            const container = document.getElementById('valuation-products');
            if (container) container.innerHTML = '<tr><td colspan="9" class="p-4 text-center text-red-500">Error: ' + error.message + '</td></tr>';
        }
    };

    window.renderValuation = function (data) {
        if (!data) data = {};
        const products = data.products || [];
        const setText = (id, text) => { const el = document.getElementById(id); if (el) el.textContent = text; };
        setText('val-total-items', data.total_items || 0);
        setText('val-total-qty', data.total_quantity || 0);
        setText('val-cost-value', 'KES ' + (data.total_cost_value || 0).toFixed(2));
        setText('val-retail-value', 'KES ' + (data.total_retail_value || 0).toFixed(2));
        setText('val-profit', 'KES ' + (data.potential_profit || 0).toFixed(2));

        // Category breakdown
        const categoriesContainer = document.getElementById('valuation-categories');
        const categories = data.categories || [];
        if (categoriesContainer) {
            if (categories.length === 0) {
                categoriesContainer.innerHTML = '<p class="text-xs text-gray-400 text-center py-2">No categories</p>';
            } else {
                let html = '';
                const totalCost = data.total_cost_value || 1;
                categories.forEach(cat => {
                    const width = Math.min((cat.total_cost_value / totalCost) * 100, 100);
                    html += '<div><div class="flex justify-between text-xs"><span class="font-medium text-gray-700">' + cat.category + '</span><span class="text-gray-500">' + cat.item_count + ' items • KES ' + cat.total_cost_value.toFixed(2) + '</span></div><div class="w-full bg-gray-200 rounded-full h-1.5 mt-0.5"><div class="bg-purple-500 h-1.5 rounded-full" style="width: ' + width + '%"></div></div></div>';
                });
                categoriesContainer.innerHTML = html;
            }
        }

        // Product table
        const productsContainer = document.getElementById('valuation-products');
        if (!productsContainer) return;
        if (products.length === 0) {
            productsContainer.innerHTML = '<tr><td colspan="9" class="p-4 text-center text-gray-500">No products in stock</td></tr>';
            return;
        }
        let html = '';
        products.forEach(p => {
            const profitClass = p.profit >= 0 ? 'text-emerald-600' : 'text-red-600';
            html += '<tr class="border-b hover:bg-purple-50/50 transition"><td class="p-2 font-medium text-gray-800">' + p.product_name + '</td><td class="p-2 font-mono text-gray-500">' + (p.barcode || 'N/A') + '</td><td class="p-2 text-gray-500">' + p.category + '</td><td class="p-2 text-center font-bold">' + p.quantity + '</td><td class="p-2 text-right font-mono">KES ' + (p.cost_price || 0).toFixed(2) + '</td><td class="p-2 text-right font-mono">KES ' + (p.retail_price || 0).toFixed(2) + '</td><td class="p-2 text-right font-mono text-amber-700">KES ' + (p.cost_value || 0).toFixed(2) + '</td><td class="p-2 text-right font-mono text-emerald-700">KES ' + (p.retail_value || 0).toFixed(2) + '</td><td class="p-2 text-right font-mono ' + profitClass + '">KES ' + (p.profit || 0).toFixed(2) + '</td></tr>';
        });
        productsContainer.innerHTML = html;
    };
    // =========================================================
    // PRODUCTS CRUD (Add / Edit)
    // =========================================================
    window.loadCategories = async function () {
        const select = document.getElementById('prodCategory');
        if (!select) return;
        select.innerHTML = '<option value="">Loading categories...</option>';
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/categories', {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            const categories = await response.json();
            select.innerHTML = '<option value="">-- Select Category --</option>';
            (Array.isArray(categories) ? categories : []).forEach(cat => {
                const opt = document.createElement('option');
                opt.value = cat.name || cat;
                opt.textContent = cat.name || cat;
                select.appendChild(opt);
            });
            const addOpt = document.createElement('option');
            addOpt.value = 'NEW';
            addOpt.textContent = '➕ Add New Category...';
            addOpt.className = 'font-semibold text-purple-800 bg-purple-50';
            select.appendChild(addOpt);
        } catch (err) {
            select.innerHTML = '<option value="">Failed to load categories</option>';
        }
    };

    window.handleCategoryChange = async function (selectElement) {
        if (selectElement.value === 'NEW') {
            const catName = prompt('Enter new category name:');
            if (catName && catName.trim() !== '') {
                try {
                    const token = window.getCookie('spide_token');
                    const res = await fetch('/api/categories/create', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            ...(token ? { 'Authorization': 'Bearer ' + token } : {})
                        },
                        body: JSON.stringify({ name: catName.trim() })
                    });
                    if (res.ok) {
                        const newCat = await res.json();
                        await window.loadCategories();
                        selectElement.value = newCat.name;
                    } else {
                        alert('Could not create category. It may already exist.');
                        selectElement.value = '';
                    }
                } catch (e) {
                    alert('Network error while adding category.');
                    selectElement.value = '';
                }
            } else {
                selectElement.value = '';
            }
        }
    };

    window.saveProduct = async function (event) {
        event.preventDefault();
        const saveBtn = document.getElementById('saveProductBtn');
        const alertBox = document.getElementById('modalAlert');
        const form = document.getElementById('addProductForm');
        const editId = form.dataset.editId;
        const isEdit = editId && editId !== '';
        const barcodeValue = document.getElementById('prodSku').value.trim();

        const payload = {
            id: isEdit ? parseInt(editId) : 0,
            barcode: barcodeValue,
            name: document.getElementById('prodName').value.trim(),
            category: document.getElementById('prodCategory').value.trim(),
            cost_price: parseFloat(document.getElementById('prodBuyingPrice').value) || 0,
            retail_price: parseFloat(document.getElementById('prodSellingPrice').value) || 0,
            wholesale_price: parseFloat(document.getElementById('prodWholesalePrice').value) || 0,
            wholesale_min_qty: parseInt(document.getElementById('prodWholesaleQty').value) || 0,
            stock_quantity: parseInt(document.getElementById('prodStockQty').value) || 0,
            reorder_level: parseInt(document.getElementById('prodReorderLevel').value) || 5
        };

        const setAlert = (msg, type) => {
            alertBox.className = 'p-3 rounded-lg text-sm font-medium ' +
                (type === 'success' ? 'bg-green-100 text-green-800 border border-green-200'
                    : type === 'warning' ? 'bg-amber-100 text-amber-800 border border-amber-200'
                        : 'bg-red-100 text-red-800 border border-red-200');
            alertBox.innerText = msg;
            alertBox.classList.remove('hidden');
        };

        if (!payload.name) { setAlert('Product name is required', 'error'); return; }
        if (payload.cost_price <= 0) { setAlert('Cost price must be > 0', 'error'); return; }
        if (payload.retail_price <= 0) { setAlert('Retail price must be > 0', 'error'); return; }
        if (payload.wholesale_price > 0 && payload.wholesale_min_qty <= 0) {
            setAlert('Wholesale min qty required when wholesale price set.', 'warning');
            return;
        }

        const url = isEdit ? '/api/products/update' : '/api/products/create';
        const method = isEdit ? 'PUT' : 'POST';
        saveBtn.disabled = true;
        saveBtn.innerText = 'Saving...';
        alertBox.classList.add('hidden');

        try {
            const token = window.getCookie('spide_token');
            const response = await fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                    ...(token ? { 'Authorization': 'Bearer ' + token } : {})
                },
                body: JSON.stringify(payload)
            });
            const result = await response.json();
            if (response.ok) {
                setAlert('Product "' + result.name + '" ' + (isEdit ? 'updated' : 'created') + '!', 'success');
                form.reset();
                const cat = document.getElementById('product-catalog-modal');
                if (cat && !cat.classList.contains('hidden') && typeof window.loadProducts === 'function') {
                    window.loadProducts();
                }
                setTimeout(() => {
                    if (typeof window.closeAddProductModal === 'function') window.closeAddProductModal();
                }, 1500);
            } else {
                setAlert('Error: ' + (result.error || 'Failed'), 'error');
            }
        } catch (error) {
            setAlert('Network error.', 'error');
        } finally {
            saveBtn.disabled = false;
            saveBtn.innerText = isEdit ? '💾 Update Product' : '💾 Save Product';
            delete form.dataset.editId;
        }
    };

    // =========================================================
    // USER MANAGEMENT
    // =========================================================
    window.loadUsers = async function () {
        const tbody = document.getElementById('user-list-body');
        if (!tbody) return;
        tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-gray-500">Loading...</td></tr>';
        try {
            const token = window.getCookie('spide_token');
            if (!token) {
                tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-red-500">Not authenticated</td></tr>';
                return;
            }
            const response = await fetch('/api/users', {
                headers: { 'Authorization': 'Bearer ' + token }
            });
            if (response.status === 401) {
                tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-red-500">Session expired.</td></tr>';
                return;
            }
            if (!response.ok) throw new Error('Failed to load users: ' + response.status);
            const data = await response.json();
            const users = Array.isArray(data) ? data : [];
            window.renderUsers(users);
        } catch (error) {
            tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-red-500">Error: ' + error.message + '</td></tr>';
        }
    };

    window.renderUsers = function (users) {
        const tbody = document.getElementById('user-list-body');
        if (!tbody) return;
        if (!users || users.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center text-gray-500">No users found</td></tr>';
            return;
        }
        let html = '';
        users.forEach(u => {
            const statusColor = u.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800';
            const statusText = u.is_active ? 'Active' : 'Inactive';
            const roleColors = {
                'admin': 'bg-purple-100 text-purple-800',
                'manager': 'bg-blue-100 text-blue-800',
                'cashier': 'bg-gray-100 text-gray-800',
                'director': 'bg-amber-100 text-amber-800'
            };
            const roleColor = roleColors[u.role] || 'bg-gray-100 text-gray-800';
            const userJSON = JSON.stringify(u).replace(/'/g, '&#39;').replace(/"/g, '&quot;');
            html += `<tr class="border-b hover:bg-gray-50/50 transition">
                <td class="p-2 font-medium text-gray-800">${u.username}</td>
                <td class="p-2">${u.full_name || '-'}</td>
                <td class="p-2 text-gray-500">${u.email || '-'}</td>
                <td class="p-2"><span class="px-2 py-0.5 rounded-full text-xs ${roleColor}">${u.role}</span></td>
                <td class="p-2 text-gray-500">${u.shop_name || 'All'}</td>
                <td class="p-2 text-center"><span class="px-2 py-0.5 rounded-full text-xs ${statusColor}">${statusText}</span></td>
                <td class="p-2 text-center space-x-1">
                    <button onclick='openEditUserModal(${userJSON})' class="text-blue-400 hover:text-blue-600 text-xs font-bold">✏️</button>
                    <button onclick="deleteUser(${u.id})" class="text-red-400 hover:text-red-600 text-xs font-bold">🗑️</button>
                </td>
            </tr>`;
        });
        tbody.innerHTML = html;
    };

    window.loadUserShops = async function () {
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/shops', {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!response.ok) return;
            const data = await response.json();
            const shopList = Array.isArray(data) ? data : [];
            const select = document.getElementById('user-shop');
            if (!select) return;
            select.innerHTML = '<option value="0">All Shops</option>';
            shopList.forEach(s => {
                const opt = document.createElement('option');
                opt.value = s.id;
                opt.textContent = s.name;
                select.appendChild(opt);
            });
        } catch (e) {
            console.error('Error loading shops:', e);
        }
    };

    window.saveUser = async function (event) {
        event.preventDefault();
        const alertBox = document.getElementById('user-alert');
        const submitBtn = document.getElementById('user-submit-btn');
        const token = window.getCookie('spide_token');
        const userId = document.getElementById('user-id').value;
        const isEdit = userId && userId !== '0';

        const data = {
            id: parseInt(userId) || 0,
            username: document.getElementById('user-username').value.trim(),
            full_name: document.getElementById('user-fullname').value.trim(),
            email: document.getElementById('user-email').value.trim(),
            role: document.getElementById('user-role').value,
            shop_id: parseInt(document.getElementById('user-shop').value) || 0
        };

        const password = document.getElementById('user-password').value;
        if (password) {
            data.password = password;
        } else if (!isEdit) {
            alertBox.className = 'p-2 rounded-lg text-sm font-medium bg-red-100 text-red-700 border border-red-200';
            alertBox.textContent = 'Password required for new users';
            alertBox.classList.remove('hidden');
            return;
        }

        if (!data.username || !data.full_name) {
            alertBox.className = 'p-2 rounded-lg text-sm font-medium bg-red-100 text-red-700 border border-red-200';
            alertBox.textContent = 'Username and Full Name required';
            alertBox.classList.remove('hidden');
            return;
        }

        submitBtn.disabled = true;
        submitBtn.textContent = 'Saving...';
        alertBox.classList.add('hidden');

        try {
            const url = isEdit ? '/api/users/update' : '/api/users/create';
            const method = isEdit ? 'PUT' : 'POST';
            const response = await fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                    ...(token ? { 'Authorization': 'Bearer ' + token } : {})
                },
                body: JSON.stringify(data)
            });
            const result = await response.json();
            if (response.ok) {
                alertBox.className = 'p-2 rounded-lg text-sm font-medium bg-green-100 text-green-700 border border-green-200';
                alertBox.textContent = '✅ User ' + (isEdit ? 'updated' : 'created') + '!';
                alertBox.classList.remove('hidden');
                setTimeout(() => {
                    window.closeUserModal();
                    window.loadUsers();
                }, 1500);
            } else {
                alertBox.className = 'p-2 rounded-lg text-sm font-medium bg-red-100 text-red-700 border border-red-200';
                alertBox.textContent = 'Error: ' + (result.error || 'Failed');
                alertBox.classList.remove('hidden');
            }
        } catch (error) {
            alertBox.className = 'p-2 rounded-lg text-sm font-medium bg-red-100 text-red-700 border border-red-200';
            alertBox.textContent = 'Network error: ' + error.message;
            alertBox.classList.remove('hidden');
        } finally {
            submitBtn.disabled = false;
            submitBtn.textContent = '💾 Save User';
        }
    };

    window.openEditUserModal = function (user) {
        if (!user) return;
        const t = document.getElementById('user-modal-title');
        if (t) t.textContent = 'Edit User';
        const setVal = (id, v) => { const el = document.getElementById(id); if (el) el.value = v; };
        setVal('user-id', user.id || 0);
        setVal('user-username', user.username || '');
        setVal('user-fullname', user.full_name || '');
        setVal('user-email', user.email || '');
        setVal('user-role', user.role || 'cashier');
        setVal('user-shop', user.shop_id || 0);
        const pw = document.getElementById('user-password');
        if (pw) { pw.value = ''; pw.required = false; }
        const hint = document.getElementById('password-hint');
        if (hint) hint.textContent = 'Leave blank to keep current password';

        // Make sure the shop dropdown is populated (in case it wasn't loaded yet)
        if (typeof window.loadUserShops === 'function') {
            window.loadUserShops().then(() => {
                const shopSelect = document.getElementById('user-shop');
                if (shopSelect) shopSelect.value = String(user.shop_id || 0);
            });
        }

        if (typeof window.SpideModals !== 'undefined' && window.SpideModals.open) {
            window.SpideModals.open('user-modal');
        } else {
            const modal = document.getElementById('user-modal');
            if (modal) modal.classList.remove('hidden');
        }
    };

    window.deleteUser = async function (id) {
        if (!confirm('Are you sure you want to delete this user?')) return;
        try {
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/users/delete?id=' + id, {
                method: 'DELETE',
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (response.ok) {
                window.loadUsers();
            } else {
                const result = await response.json();
                alert('Error: ' + (result.error || 'Failed'));
            }
        } catch (error) {
            console.error('Delete error:', error);
            alert('Network error');
        }
    };

    console.log('[SpideLoaders] ready');
})();