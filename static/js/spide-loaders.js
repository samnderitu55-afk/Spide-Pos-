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
        const printWindow = window.open('', '_blank', 'width=350,height=600');
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
        printWindow.document.write(receiptHTML);
        printWindow.document.close();
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
            const shopId = window.getCurrentShopId();
            const token = window.getCookie('spide_token');
            const response = await fetch('/api/products?shop_id=' + shopId, {
                headers: token ? { 'Authorization': 'Bearer ' + token } : {}
            });
            if (!response.ok) throw new Error('Failed: ' + response.status);
            const products = await response.json();
            window.__allProducts = products || [];
            window.__filteredProducts = [...window.__allProducts];
            window.updateCatalogCounts();
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
                    <button onclick='editProduct(${JSON.stringify(p).replace(/'/g, "&#39;")})' class="bg-blue-100 hover:bg-blue-200 text-blue-800 text-xs font-bold px-2.5 py-1 rounded-lg transition">✏️ Edit</button>
                </td>
            </tr>`;
        });
        tbody.innerHTML = html;
    };

    window.filterCatalog = function () {
        const searchInput = document.getElementById('catalog-search-input');
        if (!searchInput) return;
        const searchTerm = searchInput.value.toLowerCase().trim();
        if (!searchTerm) {
            window.__filteredProducts = [...window.__allProducts];
            const c = document.getElementById('catalog-search-count');
            if (c) c.textContent = '';
        } else {
            window.__filteredProducts = window.__allProducts.filter(p => {
                const nameMatch = p.name?.toLowerCase().includes(searchTerm) || false;
                const barcodeMatch = (p.barcode || '').toLowerCase().includes(searchTerm);
                const categoryMatch = p.category?.toLowerCase().includes(searchTerm) || false;
                return nameMatch || barcodeMatch || categoryMatch;
            });
            const countEl = document.getElementById('catalog-search-count');
            if (countEl) countEl.textContent = `${window.__filteredProducts.length} result${window.__filteredProducts.length !== 1 ? 's' : ''}`;
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
        const valuation = data.valuation || {};
        const products = data.products || [];
        const setText = (id, text) => { const el = document.getElementById(id); if (el) el.textContent = text; };
        setText('val-total-items', valuation.total_items || 0);
        setText('val-total-qty', valuation.total_quantity || 0);
        setText('val-cost-value', 'KES ' + (valuation.total_cost_value || 0).toFixed(2));
        setText('val-retail-value', 'KES ' + (valuation.total_retail_value || 0).toFixed(2));
        setText('val-profit', 'KES ' + (valuation.potential_profit || 0).toFixed(2));
        const categoriesContainer = document.getElementById('valuation-categories');
        const categories = valuation.categories || [];
        if (categoriesContainer) {
            if (categories.length === 0) {
                categoriesContainer.innerHTML = '<p class="text-xs text-gray-400 text-center py-2">No categories</p>';
            } else {
                let html = '';
                const totalCost = valuation.total_cost_value || 1;
                categories.forEach(cat => {
                    const width = Math.min((cat.total_cost_value / totalCost) * 100, 100);
                    html += '<div><div class="flex justify-between text-xs"><span class="font-medium text-gray-700">' + cat.category + '</span><span class="text-gray-500">' + cat.item_count + ' items • KES ' + cat.total_cost_value.toFixed(2) + '</span></div><div class="w-full bg-gray-200 rounded-full h-1.5 mt-0.5"><div class="bg-purple-500 h-1.5 rounded-full" style="width: ' + width + '%"></div></div></div>';
                });
                categoriesContainer.innerHTML = html;
            }
        }
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

    console.log('[SpideLoaders] ready');
})();