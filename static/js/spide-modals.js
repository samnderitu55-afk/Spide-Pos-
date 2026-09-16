/* =========================================================
   Spide POS — Shared Modals
   Load on every page: <script src="/static/js/spide-modals.js"></script>
   Exposes: window.SpideModals.open(id), .close(id)
   Also defines every window.openXModal / closeXModal used by the HTML.
   ========================================================= */

(function () {
    'use strict';

    // ---------------------------------------------------------
    // 1. MODAL HTML — paste the entire contents of modals.html here
    // ---------------------------------------------------------
    const MODAL_HTML = `
        <!-- ============================================ -->
        <!-- ADD PRODUCT MODAL -->
        <!-- ============================================ -->
        <div id="addProductModal"
            class="fixed inset-0 bg-gray-900/60 backdrop-blur-sm hidden items-center justify-center z-50 p-4 overflow-y-auto">
            <div class="bg-white rounded-2xl shadow-2xl w-full max-w-2xl border border-gray-100 transform transition-all my-8">
                <div class="bg-purple-900 text-white px-6 py-4 rounded-t-2xl flex justify-between items-center">
                    <h3 class="text-lg font-bold flex items-center gap-2">📦 Add New Stock Item</h3>
                    <button onclick="closeAddProductModal()"
                        class="text-purple-200 hover:text-white hover:bg-purple-800 p-1 rounded-lg text-xl font-bold transition">✕</button>
                </div>
                <form id="addProductForm" onsubmit="saveProduct(event)" class="p-6 space-y-4">
                    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 uppercase mb-1">SKU / Barcode *</label>
                            <input type="text" id="prodSku" required placeholder="e.g. COSM-001"
                                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-600 focus:border-transparent outline-none text-sm">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 uppercase mb-1">Product Name *</label>
                            <input type="text" id="prodName" required placeholder="e.g. Matte Lipstick - Ruby Red"
                                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-600 focus:border-transparent outline-none text-sm">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 uppercase mb-1">Category *</label>
                            <select id="prodCategory" required onchange="handleCategoryChange(this)"
                                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-600 focus:border-transparent outline-none text-sm bg-white">
                                <option value="">-- Select Category --</option>
                            </select>
                        </div>
                    </div>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 uppercase mb-1">Buying Price (KES) *</label>
                            <input type="number" step="0.01" id="prodBuyingPrice" required placeholder="0.00"
                                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-600 focus:border-transparent outline-none text-sm">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 uppercase mb-1">Selling Price (Retail)
                                *</label>
                            <input type="number" step="0.01" id="prodSellingPrice" required placeholder="0.00"
                                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-600 focus:border-transparent outline-none text-sm">
                        </div>
                    </div>
                    <div class="bg-purple-50/60 p-4 rounded-xl border border-purple-100 grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label class="block text-xs font-semibold text-purple-900 uppercase mb-1">Wholesale Price
                                (KES)</label>
                            <input type="number" step="0.01" id="prodWholesalePrice" placeholder="0.00"
                                class="w-full px-3 py-2 border border-purple-200 rounded-lg focus:ring-2 focus:ring-purple-600 outline-none text-sm bg-white">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-purple-900 uppercase mb-1">Wholesale Min. Qty</label>
                            <input type="number" id="prodWholesaleQty" placeholder="e.g. 12"
                                class="w-full px-3 py-2 border border-purple-200 rounded-lg focus:ring-2 focus:ring-purple-600 outline-none text-sm bg-white">
                        </div>
                    </div>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 uppercase mb-1">Initial Stock Qty *</label>
                            <input type="number" id="prodStockQty" required placeholder="0"
                                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-600 focus:border-transparent outline-none text-sm">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 uppercase mb-1">Re-Order Level Alert
                                *</label>
                            <input type="number" id="prodReorderLevel" value="5" required placeholder="5"
                                class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-600 focus:border-transparent outline-none text-sm">
                        </div>
                    </div>
                    <div id="modalAlert" class="hidden p-3 rounded-lg text-sm font-medium"></div>
                    <div class="flex justify-end gap-3 pt-4 border-t border-gray-100">
                        <button type="button" onclick="closeAddProductModal()"
                            class="px-4 py-2 text-sm font-semibold text-gray-600 bg-gray-100 hover:bg-gray-200 rounded-lg transition">Cancel</button>
                        <button type="submit" id="saveProductBtn"
                            class="px-5 py-2 text-sm font-semibold text-white bg-purple-800 hover:bg-purple-700 rounded-lg shadow-md transition flex items-center gap-2">💾
                            Save Product</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- ============================================ -->
        <!-- PRODUCT CATALOG MODAL -->
        <!-- ============================================ -->
        <div id="product-catalog-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-6xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b pb-3">
                    <div>
                        <h2 class="text-xl font-bold text-gray-900 flex items-center gap-2">📦 Product Catalog</h2>
                        <p class="text-xs text-gray-500">Complete list of products in inventory</p>
                    </div>
                    <div class="flex items-center gap-2">
                        <button onclick="loadProducts()"
                            class="bg-purple-100 hover:bg-purple-200 text-purple-800 text-xs font-bold px-3 py-1.5 rounded-lg transition flex items-center gap-1">🔄
                            Refresh</button>
                        <button onclick="closeProductCatalogModal()"
                            class="text-gray-400 hover:text-gray-600 font-bold text-xl ml-2">✕</button>
                    </div>
                </div>

                <!-- ✅ SEARCH BAR -->
                <div class="flex items-center gap-3 bg-gray-50 p-3 rounded-xl border border-gray-200">
                    <div class="flex-1 relative">
                        <input type="text" id="catalog-search-input" placeholder="🔍 Search by product name or barcode..."
                            class="w-full px-4 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none bg-white"
                            oninput="filterCatalog()">
                        <span id="catalog-search-count"
                            class="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-gray-400"></span>
                    </div>
                    <button onclick="clearCatalogSearch()"
                        class="text-gray-400 hover:text-gray-600 text-sm font-medium px-2 py-1">
                        ✕ Clear
                    </button>
                </div>

                <div class="overflow-y-auto flex-1 border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-sm">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-3">Barcode / SKU</th>
                                <th class="p-3">Product Name</th>
                                <th class="p-3">Category</th>
                                <th class="p-3 text-right">Retail Price</th>
                                <th class="p-3 text-right">Wholesale Price</th>
                                <th class="p-3 text-center">Stock</th>
                                <th class="p-3 text-center">Actions</th>
                            </tr>
                        </thead>
                        <tbody id="product-catalog-rows">
                            <tr>
                                <td colspan="7" class="p-4 text-center text-gray-500">Loading products...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div class="flex justify-between items-center border-t pt-3">
                    <span id="catalog-total-count" class="text-xs text-gray-400"></span>
                    <button onclick="closeProductCatalogModal()"
                        class="bg-gray-200 text-gray-800 font-semibold px-5 py-2 rounded-xl hover:bg-gray-300 transition">Close</button>
                </div>
            </div>
        </div>

        <!-- ============================================ -->
        <!-- LOW STOCK MODAL -->
        <!-- ============================================ -->
        <div id="low-stock-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-5xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b pb-3">
                    <div>
                        <h2 class="text-xl font-bold text-gray-900 flex items-center gap-2"><span
                                class="text-amber-500">⚠️</span> Low Stock & Reorder Alerts</h2>
                        <p class="text-xs text-gray-500">Products requiring immediate supplier reordering</p>
                    </div>
                    <button onclick="closeLowStockModal()"
                        class="text-gray-400 hover:text-gray-600 font-bold text-xl">✕</button>
                </div>
                <div id="low-stock-content" class="overflow-y-auto flex-1 border rounded-xl relative p-4">
                    <p class="text-center text-gray-500 py-6">Loading stock status...</p>
                </div>
                <div class="flex justify-end border-t pt-3">
                    <button onclick="closeLowStockModal()"
                        class="bg-gray-200 text-gray-800 font-semibold px-5 py-2 rounded-xl hover:bg-gray-300 transition">Close</button>
                </div>
            </div>
        </div>

        <!-- INVENTORY VALUATION MODAL -->
        <div id="inventory-valuation-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-6xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b pb-3">
                    <div>
                        <h2 class="text-xl font-bold text-gray-900 flex items-center gap-2">💎 Inventory Valuation</h2>
                        <p class="text-xs text-gray-500">Total value of inventory across all shops</p>
                    </div>
                    <div class="flex items-center gap-2">
                        <select id="valuation-branch-select" onchange="loadInventoryValuation()"
                            class="text-xs border rounded-lg px-2 py-1 bg-white">
                            <option value="0">All Shops</option>
                        </select>
                        <button onclick="loadInventoryValuation()"
                            class="bg-purple-100 hover:bg-purple-200 text-purple-800 text-xs font-bold px-3 py-1.5 rounded-lg transition">🔄
                            Refresh</button>
                        <button onclick="closeInventoryValuationModal()"
                            class="text-gray-400 hover:text-gray-600 font-bold text-xl">✕</button>
                    </div>
                </div>

                <!-- Summary Cards -->
                <div class="grid grid-cols-2 md:grid-cols-5 gap-3">
                    <div class="bg-purple-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">Total Items</p>
                        <p id="val-total-items" class="text-xl font-bold text-purple-900">0</p>
                    </div>
                    <div class="bg-blue-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">Total Quantity</p>
                        <p id="val-total-qty" class="text-xl font-bold text-blue-700">0</p>
                    </div>
                    <div class="bg-amber-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">Cost Value</p>
                        <p id="val-cost-value" class="text-xl font-bold text-amber-700">KES 0.00</p>
                    </div>
                    <div class="bg-emerald-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">Retail Value</p>
                        <p id="val-retail-value" class="text-xl font-bold text-emerald-700">KES 0.00</p>
                    </div>
                    <div class="bg-purple-100 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">Potential Profit</p>
                        <p id="val-profit" class="text-xl font-bold text-purple-900">KES 0.00</p>
                    </div>
                </div>

                <!-- Category Breakdown -->
                <div class="border rounded-xl p-3 max-h-48 overflow-y-auto">
                    <h3 class="font-semibold text-gray-800 text-sm mb-2">📊 Category Breakdown</h3>
                    <div id="valuation-categories" class="space-y-1">
                        <p class="text-xs text-gray-400 text-center py-2">Loading categories...</p>
                    </div>
                </div>

                <!-- Product List -->
                <div class="flex-1 overflow-y-auto border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-xs">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-2">Product</th>
                                <th class="p-2">Barcode</th>
                                <th class="p-2">Category</th>
                                <th class="p-2 text-center">Qty</th>
                                <th class="p-2 text-right">Cost Price</th>
                                <th class="p-2 text-right">Retail Price</th>
                                <th class="p-2 text-right">Cost Value</th>
                                <th class="p-2 text-right">Retail Value</th>
                                <th class="p-2 text-right">Profit</th>
                            </tr>
                        </thead>
                        <tbody id="valuation-products">
                            <tr>
                                <td colspan="9" class="p-4 text-center text-gray-500">Loading products...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div class="flex justify-end border-t pt-3">
                    <button onclick="closeInventoryValuationModal()"
                        class="bg-gray-200 text-gray-800 font-semibold px-5 py-2 rounded-xl hover:bg-gray-300 transition">Close</button>
                </div>
            </div>
        </div>

        <!-- Z-REPORT MODAL - With Role-Based Shop Filter -->
        <div id="zreport-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-4xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <!-- Header -->
                <div class="flex justify-between items-center border-b border-gray-100 pb-4">
                    <div>
                        <h2 class="text-xl font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-purple-100 p-1.5 rounded-lg">📊</span>
                            Daily Z-Report
                        </h2>
                        <p class="text-xs text-gray-500 mt-0.5">Complete sales summary for the day</p>
                    </div>
                    <div class="flex items-center gap-2 flex-wrap">
                        <!-- Shop Filter Container - Only visible to Directors/Admins -->
                        <div id="zreport-shop-filter-container"
                            class="hidden items-center gap-1 bg-gray-50 rounded-lg px-2 py-1 border border-gray-200">
                            <span class="text-xs text-gray-500">🏪</span>
                            <select id="zreport-shop-filter" onchange="loadZReport()"
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1">
                                <option value="0">All Shops</option>
                            </select>
                        </div>
                        <div class="flex items-center gap-1 bg-gray-50 rounded-lg px-2 py-1 border border-gray-200">
                            <span class="text-xs text-gray-500">📅</span>
                            <input type="date" id="zreport-date" value=""
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1 w-32">
                        </div>
                        <button onclick="loadZReport()"
                            class="bg-purple-600 hover:bg-purple-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition flex items-center gap-1">
                            <span>⟳</span> Generate
                        </button>
                        <button onclick="printZReport()"
                            class="bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition flex items-center gap-1">
                            🖨️ Print
                        </button>
                        <button onclick="closeZReportModal()"
                            class="text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg p-1.5 transition text-lg">
                            ✕
                        </button>
                    </div>
                </div>

                <!-- Report Content -->
                <div id="zreport-content" class="flex-1 overflow-y-auto">
                    <div class="text-center text-gray-400 py-12 flex flex-col items-center">
                        <span class="text-4xl mb-3">📋</span>
                        <p class="text-sm font-medium">Select a date and shop, then click Generate</p>
                        <p class="text-xs mt-1">View daily sales, payments, expenses, and profit</p>
                    </div>
                </div>

                <!-- Footer -->
                <div class="flex justify-end border-t border-gray-100 pt-3">
                    <button onclick="closeZReportModal()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">
                        Close
                    </button>
                </div>
            </div>
        </div>

        <!-- PRODUCT SALES & COGS REPORT MODAL -->
        <div id="product-sales-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-6xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <!-- Header -->
                <div class="flex justify-between items-center border-b border-gray-100 pb-4">
                    <div>
                        <h2 class="text-xl font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-blue-100 p-1.5 rounded-lg">📈</span>
                            Product Sales &amp; COGS
                        </h2>
                        <p class="text-xs text-gray-500 mt-0.5">Sales, cost of goods sold, and profit by product</p>
                    </div>
                    <div class="flex items-center gap-2">
                        <div class="flex items-center gap-1 bg-gray-50 rounded-lg px-2 py-1 border border-gray-200">
                            <span class="text-xs text-gray-500">📅</span>
                            <input type="date" id="productsales-start-date" value=""
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1 w-28">
                            <span class="text-xs text-gray-400">to</span>
                            <input type="date" id="productsales-end-date" value=""
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1 w-28">
                        </div>
                        <div id="productsales-shop-filter-container" class="hidden items-center gap-1 bg-gray-50 rounded-lg px-2 py-1 border border-gray-200">
                            <span class="text-xs text-gray-500">🏪</span>
                            <select id="productsales-shop-filter"
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1">
                                <option value="0">All Shops</option>
                            </select>
                        </div>
                        <!-- Director-only category filter -->
                        <div id="productsales-category-filter-container"
                            class="hidden items-center gap-1 bg-gray-50 rounded-lg px-2 py-1 border border-gray-200">
                            <span class="text-xs text-gray-500">🏷️</span>
                            <select id="productsales-category-filter" onchange="onProductSalesCategoryChange()"
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1 max-w-[140px]">
                                <option value="">All Categories</option>
                            </select>
                        </div>

                        <!-- Director-only product filter -->
                        <div id="productsales-product-filter-container"
                            class="hidden items-center gap-1 bg-gray-50 rounded-lg px-2 py-1 border border-gray-200">
                            <span class="text-xs text-gray-500">📦</span>
                            <select id="productsales-product-filter" onchange="loadProductSalesReport()"
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1 max-w-[180px]">
                                <option value="">All Products</option>
                            </select>
                        </div>
                        <button onclick="loadProductSalesReport()"
                            class="bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition flex items-center gap-1">
                            <span>⟳</span> Generate
                        </button>
                        <button onclick="printProductSalesReport()"
                            class="bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition flex items-center gap-1">
                            🖨️ Print
                        </button>
                        <button onclick="closeProductSalesModal()"
                            class="text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg p-1.5 transition text-lg">
                            ✕
                        </button>
                    </div>
                </div>

                <!-- Summary Stats -->
                <div class="grid grid-cols-2 md:grid-cols-4 gap-3" id="productsales-summary">
                    <div
                        class="bg-gradient-to-br from-blue-50 to-blue-100/50 p-3 rounded-xl text-center border border-blue-200/50">
                        <p class="text-xs font-medium text-blue-600 uppercase tracking-wider">Total Revenue</p>
                        <p id="ps-total-revenue" class="text-lg font-bold text-blue-900">KES 0.00</p>
                    </div>
                    <div
                        class="bg-gradient-to-br from-amber-50 to-amber-100/50 p-3 rounded-xl text-center border border-amber-200/50">
                        <p class="text-xs font-medium text-amber-600 uppercase tracking-wider">Total COGS</p>
                        <p id="ps-total-cogs" class="text-lg font-bold text-amber-900">KES 0.00</p>
                    </div>
                    <div
                        class="bg-gradient-to-br from-emerald-50 to-emerald-100/50 p-3 rounded-xl text-center border border-emerald-200/50">
                        <p class="text-xs font-medium text-emerald-600 uppercase tracking-wider">Gross Profit</p>
                        <p id="ps-total-profit" class="text-lg font-bold text-emerald-900">KES 0.00</p>
                    </div>
                    <div
                        class="bg-gradient-to-br from-purple-50 to-purple-100/50 p-3 rounded-xl text-center border border-purple-200/50">
                        <p class="text-xs font-medium text-purple-600 uppercase tracking-wider">Avg Margin</p>
                        <p id="ps-avg-margin" class="text-lg font-bold text-purple-900">0%</p>
                    </div>
                </div>

                <!-- Product List -->
                <div class="flex-1 overflow-y-auto border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-xs">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-2">Product</th>
                                <th class="p-2">Category</th>
                                <th class="p-2 text-center">Units Sold</th>
                                <th class="p-2 text-right">Total Revenue</th>
                                <th class="p-2 text-right">Total COGS</th>
                                <th class="p-2 text-right">Gross Profit</th>
                                <th class="p-2 text-right">Margin %</th>
                            </tr>
                        </thead>
                        <tbody id="productsales-table-body">
                            <tr>
                                <td colspan="7" class="p-4 text-center text-gray-500">Select date range and click Generate
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <!-- Footer -->
                <div class="flex justify-end border-t border-gray-100 pt-3">
                    <button onclick="closeProductSalesModal()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">
                        Close
                    </button>
                </div>
            </div>
        </div>

        <!-- ADD EXPENSE MODAL -->
        <div id="expense-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-lg rounded-2xl shadow-2xl p-6 space-y-4">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-amber-100 p-1.5 rounded-lg">💰</span>
                        Add Expense
                    </h2>
                    <button onclick="closeExpenseModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <form id="expense-form" onsubmit="submitExpense(event)">
                    <div class="space-y-3">
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Category *</label>
                            <select id="expense-category" required
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-amber-500 focus:outline-none">
                                <option value="">Select Category</option>
                                <option value="Rent">Rent</option>
                                <option value="Utilities">Utilities</option>
                                <option value="Staff Food">Staff Food</option>
                                <option value="Salaries">Salaries</option>
                                <option value="Supplies">Supplies</option>
                                <option value="Maintenance">Maintenance</option>
                                <option value="Transport">Transport</option>
                                <option value="Marketing">Marketing</option>
                                <option value="Insurance">Insurance</option>
                                <option value="Taxes">Taxes</option>
                                <option value="Other">Other</option>
                            </select>
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Description *</label>
                            <input type="text" id="expense-description" required placeholder="What was this expense for?"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-amber-500 focus:outline-none">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Amount (KES) *</label>
                            <input type="number" id="expense-amount" required step="0.01" min="0.01" placeholder="0.00"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-amber-500 focus:outline-none">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Date</label>
                            <input type="date" id="expense-date"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-amber-500 focus:outline-none">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Payment Method</label>
                            <select id="expense-payment-method"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-amber-500 focus:outline-none">
                                <option value="Cash">Cash</option>
                                <option value="M-Pesa">M-Pesa</option>
                                <option value="Bank Transfer">Bank Transfer</option>
                                <option value="Card">Card</option>
                                <option value="Other">Other</option>
                            </select>
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Reference (Optional)</label>
                            <input type="text" id="expense-reference" placeholder="Receipt # or reference"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-amber-500 focus:outline-none">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Notes (Optional)</label>
                            <textarea id="expense-notes" rows="2" placeholder="Any additional notes..."
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-amber-500 focus:outline-none"></textarea>
                        </div>
                        <div id="expense-alert" class="hidden p-2 rounded-lg text-sm font-medium"></div>
                    </div>
                    <div class="flex gap-3 pt-4 border-t border-gray-100 mt-4">
                        <button type="button" onclick="closeExpenseModal()"
                            class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">Cancel</button>
                        <button type="submit" id="expense-submit-btn"
                            class="flex-1 bg-amber-600 hover:bg-amber-700 text-white font-medium py-2 rounded-lg transition text-sm">💾
                            Save Expense</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- EXPENSE REPORT MODAL -->
        <div id="expense-report-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-5xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <div>
                        <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-amber-100 p-1.5 rounded-lg">📊</span>
                            Expense Report
                        </h2>
                        <p class="text-xs text-gray-500">View and manage all expenses</p>
                    </div>
                    <div class="flex items-center gap-2">
                        <div class="flex items-center gap-1 bg-gray-50 rounded-lg px-2 py-1 border border-gray-200">
                            <span class="text-xs text-gray-500">📅</span>
                            <input type="date" id="expense-report-start"
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1 w-28">
                            <span class="text-xs text-gray-400">to</span>
                            <input type="date" id="expense-report-end"
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1 w-28">
                        </div>
                        <select id="expense-report-category" class="text-xs border rounded-lg px-2 py-1.5 bg-white">
                            <option value="">All Categories</option>
                        </select>
                        <button onclick="loadExpenseReport()"
                            class="bg-amber-600 hover:bg-amber-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition">⟳
                            Refresh</button>
                        <button onclick="closeExpenseReportModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                    </div>
                </div>

                <!-- Summary -->
                <div class="grid grid-cols-2 gap-3">
                    <div class="bg-amber-50 p-3 rounded-xl text-center border border-amber-200/50">
                        <p class="text-xs font-medium text-amber-600 uppercase tracking-wider">Total Expenses</p>
                        <p id="expense-report-total" class="text-xl font-bold text-amber-900">KES 0.00</p>
                    </div>
                    <div class="bg-blue-50 p-3 rounded-xl text-center border border-blue-200/50">
                        <p class="text-xs font-medium text-blue-600 uppercase tracking-wider">Number of Expenses</p>
                        <p id="expense-report-count" class="text-xl font-bold text-blue-900">0</p>
                    </div>
                </div>

                <!-- Expense List -->
                <div class="flex-1 overflow-y-auto border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-xs">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-2">Date</th>
                                <th class="p-2">Category</th>
                                <th class="p-2">Description</th>
                                <th class="p-2 text-right">Amount</th>
                                <th class="p-2">Payment</th>
                                <th class="p-2">Branch</th>
                                <th class="p-2 text-center">Actions</th>
                            </tr>
                        </thead>
                        <tbody id="expense-report-body">
                            <tr>
                                <td colspan="7" class="p-4 text-center text-gray-500">Select date range and click Refresh
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div class="flex justify-end border-t border-gray-100 pt-3">
                    <button onclick="closeExpenseReportModal()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">Close</button>
                </div>
            </div>
        </div>

        <!-- NEW TRANSFER MODAL -->
        <div id="transfer-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-2xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-blue-100 p-1.5 rounded-lg">📦</span>
                        New Stock Transfer
                    </h2>
                    <button onclick="closeTransferModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <form id="transfer-form" onsubmit="submitTransfer(event)" class="flex-1 flex flex-col">
                    <div class="space-y-3">
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Destination Shop *</label>
                            <select id="transfer-to-shop" required
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none">
                                <option value="">Select Shop</option>
                            </select>
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Transfer Date</label>
                            <input type="date" id="transfer-date"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Notes (Optional)</label>
                            <textarea id="transfer-notes" rows="2" placeholder="Reason for transfer..."
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"></textarea>
                        </div>

                        <!-- Transfer Items Section -->
                        <div class="border-t pt-3">
                            <h3 class="text-sm font-semibold text-gray-700 mb-2">📋 Items to Transfer</h3>
                            <div id="transfer-items-container" class="space-y-2">
                                <div class="flex gap-2 items-center">
                                    <div class="flex-1 relative">
                                        <input type="text" id="transfer-product-search"
                                            placeholder="Search by barcode or product name..."
                                            class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
                                            autocomplete="off" oninput="handleTransferSearch(this.value)"
                                            onkeydown="handleTransferSearchKeydown(event)">
                                        <div id="transfer-search-results"
                                            class="absolute left-0 right-0 top-full mt-1 bg-white border border-gray-200 rounded-lg shadow-xl hidden z-50 max-h-48 overflow-y-auto">
                                        </div>
                                    </div>
                                    <input type="number" id="transfer-qty" placeholder="Qty"
                                        class="w-20 px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
                                        value="1" min="1">
                                    <button type="button" onclick="addTransferItem()"
                                        class="bg-blue-600 hover:bg-blue-700 text-white px-3 py-2 rounded-lg text-sm whitespace-nowrap">➕
                                        Add</button>
                                </div>
                            </div>
                            <div id="transfer-items-list" class="mt-2 space-y-1 max-h-32 overflow-y-auto">
                                <p class="text-xs text-gray-400 text-center py-2">No items added</p>
                            </div>
                        </div>

                        <div id="transfer-alert" class="hidden p-2 rounded-lg text-sm font-medium"></div>
                    </div>
                    <div class="flex gap-3 pt-4 border-t border-gray-100 mt-4">
                        <button type="button" onclick="closeTransferModal()"
                            class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">Cancel</button>
                        <button type="submit" id="transfer-submit-btn"
                            class="flex-1 bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 rounded-lg transition text-sm">📦
                            Create Transfer</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- TRANSFER HISTORY MODAL -->
        <div id="transfer-history-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-5xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <div>
                        <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-blue-100 p-1.5 rounded-lg">📋</span>
                            Transfer History
                        </h2>
                        <p class="text-xs text-gray-500">View all stock transfers</p>
                    </div>
                    <button onclick="closeTransferHistoryModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <div class="flex-1 overflow-y-auto border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-xs">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-2">Transfer #</th>
                                <th class="p-2">From</th>
                                <th class="p-2">To</th>
                                <th class="p-2 text-center">Items</th>
                                <th class="p-2 text-right">Total Cost</th>
                                <th class="p-2">Date</th>
                                <th class="p-2">Status</th>
                                <th class="p-2 text-center">Actions</th>
                            </tr>
                        </thead>
                        <tbody id="transfer-history-body">
                            <tr>
                                <td colspan="8" class="p-4 text-center text-gray-500">Loading transfers...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div class="flex justify-end border-t border-gray-100 pt-3">
                    <button onclick="closeTransferHistoryModal()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">Close</button>
                </div>
            </div>
        </div>

        <!-- TRANSFER DETAIL MODAL -->
        <div id="transfer-detail-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-3xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <div>
                        <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-blue-100 p-1.5 rounded-lg">📄</span>
                            Transfer Details
                        </h2>
                        <p id="transfer-detail-number" class="text-xs text-gray-500">#TRF-XXXXXXXX</p>
                    </div>
                    <button onclick="closeTransferDetailModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <!-- Transfer Info -->
                <div class="grid grid-cols-2 md:grid-cols-4 gap-3" id="transfer-detail-info">
                    <div class="bg-gray-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">From</p>
                        <p id="td-from-shop" class="font-semibold text-gray-800">-</p>
                    </div>
                    <div class="bg-gray-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">To</p>
                        <p id="td-to-shop" class="font-semibold text-gray-800">-</p>
                    </div>
                    <div class="bg-gray-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">Date</p>
                        <p id="td-date" class="font-semibold text-gray-800">-</p>
                    </div>
                    <div class="bg-gray-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">Status</p>
                        <p id="td-status" class="font-semibold text-gray-800">-</p>
                    </div>
                </div>

                <!-- Items List -->
                <div class="flex-1 overflow-y-auto border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-xs">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-2">Product</th>
                                <th class="p-2">Barcode</th>
                                <th class="p-2 text-center">Quantity</th>
                                <th class="p-2 text-right">Cost Price</th>
                                <th class="p-2 text-right">Subtotal</th>
                            </tr>
                        </thead>
                        <tbody id="transfer-detail-items">
                            <tr>
                                <td colspan="5" class="p-4 text-center text-gray-500">Loading items...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <!-- Footer -->
                <div class="flex justify-between items-center border-t border-gray-100 pt-3">
                    <div>
                        <span class="text-xs text-gray-500">Total Items:</span>
                        <span id="td-total-items" class="font-bold text-gray-800 ml-1">0</span>
                        <span class="text-xs text-gray-500 ml-3">Total Cost:</span>
                        <span id="td-total-cost" class="font-bold text-amber-700 ml-1">KES 0.00</span>
                    </div>
                    <button onclick="closeTransferDetailModal()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">Close</button>
                </div>
            </div>
        </div>

        <!-- NEW PURCHASE MODAL -->
        <div id="purchase-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-2xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-emerald-100 p-1.5 rounded-lg">📦</span>
                        New Purchase Order
                    </h2>
                    <button onclick="closePurchaseModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <form id="purchase-form" onsubmit="submitPurchase(event)" class="flex-1 flex flex-col">
                    <div class="space-y-3">
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Supplier *</label>
                            <select id="purchase-supplier" required
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none">
                                <option value="">Select Supplier</option>
                            </select>
                            <button type="button" onclick="openSupplierModal()"
                                class="text-xs text-emerald-600 hover:text-emerald-800 mt-1">➕ Add New Supplier</button>
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Purchase Date</label>
                            <input type="date" id="purchase-date"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none">
                        </div>
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">Notes (Optional)</label>
                            <textarea id="purchase-notes" rows="2" placeholder="Additional notes..."
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none"></textarea>
                        </div>

                        <!-- Purchase Items Section -->
                        <div class="border-t pt-3">
                            <h3 class="text-sm font-semibold text-gray-700 mb-2">📋 Items to Purchase</h3>
                            <div id="purchase-items-container" class="space-y-2">
                                <div class="flex gap-2 items-center">
                                    <div class="flex-1 relative">
                                        <input type="text" id="purchase-product-search"
                                            placeholder="Search by barcode or product name..."
                                            class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none"
                                            autocomplete="off" oninput="handlePurchaseSearch(this.value)"
                                            onkeydown="handlePurchaseSearchKeydown(event)">
                                        <div id="purchase-search-results"
                                            class="absolute left-0 right-0 top-full mt-1 bg-white border border-gray-200 rounded-lg shadow-xl hidden z-50 max-h-48 overflow-y-auto">
                                        </div>
                                    </div>
                                    <input type="number" id="purchase-qty" placeholder="Qty"
                                        class="w-20 px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none"
                                        value="1" min="1">
                                    <input type="number" id="purchase-cost" placeholder="Cost"
                                        class="w-28 px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none"
                                        value="0.00" step="0.01">
                                    <button type="button" onclick="addPurchaseItem()"
                                        class="bg-emerald-600 hover:bg-emerald-700 text-white px-3 py-2 rounded-lg text-sm whitespace-nowrap">➕
                                        Add</button>
                                </div>
                            </div>
                            <div id="purchase-items-list" class="mt-2 space-y-1 max-h-32 overflow-y-auto">
                                <p class="text-xs text-gray-400 text-center py-2">No items added</p>
                            </div>
                        </div>

                        <div id="purchase-alert" class="hidden p-2 rounded-lg text-sm font-medium"></div>
                    </div>
                    <div class="flex gap-3 pt-4 border-t border-gray-100 mt-4">
                        <button type="button" onclick="closePurchaseModal()"
                            class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">Cancel</button>
                        <button type="submit" id="purchase-submit-btn"
                            class="flex-1 bg-emerald-600 hover:bg-emerald-700 text-white font-medium py-2 rounded-lg transition text-sm">📦
                            Create Purchase</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- NEW SUPPLIER MODAL -->
        <div id="supplier-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-md rounded-2xl shadow-2xl p-6 space-y-4">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-emerald-100 p-1.5 rounded-lg">👥</span>
                        Add Supplier
                    </h2>
                    <button onclick="closeSupplierModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <form id="supplier-form" onsubmit="submitSupplier(event)">
                    <div class="space-y-3">
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Supplier Name *</label><input
                                type="text" id="supplier-name" required placeholder="e.g. ABC Distributors"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Contact Person</label><input
                                type="text" id="supplier-contact" placeholder="e.g. John Doe"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Phone</label><input type="text"
                                id="supplier-phone" placeholder="e.g. 0712345678"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Email</label><input type="email"
                                id="supplier-email" placeholder="e.g. info@abc.co.ke"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Address</label><textarea
                                id="supplier-address" rows="2" placeholder="Physical address..."
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none"></textarea>
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Notes</label><textarea
                                id="supplier-notes" rows="2" placeholder="Additional notes..."
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none"></textarea>
                        </div>
                        <div id="supplier-alert" class="hidden p-2 rounded-lg text-sm font-medium"></div>
                    </div>
                    <div class="flex gap-3 pt-4 border-t border-gray-100 mt-4">
                        <button type="button" onclick="closeSupplierModal()"
                            class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">Cancel</button>
                        <button type="submit" id="supplier-submit-btn"
                            class="flex-1 bg-emerald-600 hover:bg-emerald-700 text-white font-medium py-2 rounded-lg transition text-sm">💾
                            Save Supplier</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- PURCHASE REPORT MODAL -->
        <div id="purchase-report-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-5xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <div>
                        <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-emerald-100 p-1.5 rounded-lg">📊</span>
                            Purchase Report
                        </h2>
                        <p class="text-xs text-gray-500">View all purchase orders</p>
                    </div>
                    <div class="flex items-center gap-2">
                        <div class="flex items-center gap-1 bg-gray-50 rounded-lg px-2 py-1 border border-gray-200">
                            <span class="text-xs text-gray-500">📅</span>
                            <input type="date" id="purchase-report-start"
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1 w-28">
                            <span class="text-xs text-gray-400">to</span>
                            <input type="date" id="purchase-report-end"
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1 w-28">
                        </div>
                        <select id="purchase-report-supplier" class="text-xs border rounded-lg px-2 py-1.5 bg-white">
                            <option value="0">All Suppliers</option>
                        </select>
                        <button onclick="loadPurchaseReport()"
                            class="bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition">⟳
                            Refresh</button>
                        <button onclick="closePurchaseReportModal()"
                            class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                    </div>
                </div>

                <!-- Summary -->
                <div class="grid grid-cols-2 gap-3">
                    <div class="bg-emerald-50 p-3 rounded-xl text-center border border-emerald-200/50">
                        <p class="text-xs font-medium text-emerald-600 uppercase tracking-wider">Total Purchases</p>
                        <p id="purchase-report-total" class="text-xl font-bold text-emerald-900">KES 0.00</p>
                    </div>
                    <div class="bg-blue-50 p-3 rounded-xl text-center border border-blue-200/50">
                        <p class="text-xs font-medium text-blue-600 uppercase tracking-wider">Number of Orders</p>
                        <p id="purchase-report-count" class="text-xl font-bold text-blue-900">0</p>
                    </div>
                </div>

                <!-- Purchase List -->
                <div class="flex-1 overflow-y-auto border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-xs">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-2">PO #</th>
                                <th class="p-2">Supplier</th>
                                <th class="p-2 text-center">Items</th>
                                <th class="p-2 text-right">Total Cost</th>
                                <th class="p-2">Date</th>
                                <th class="p-2">Created By</th>
                                <th class="p-2 text-center">Actions</th>
                            </tr>
                        </thead>
                        <tbody id="purchase-report-body">
                            <tr>
                                <td colspan="7" class="p-4 text-center text-gray-500">Select date range and click Refresh
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div class="flex justify-end border-t border-gray-100 pt-3">
                    <button onclick="closePurchaseReportModal()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">Close</button>
                </div>
            </div>
        </div>

        <!-- PURCHASE DETAIL MODAL -->
        <div id="purchase-detail-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-3xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <div>
                        <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-emerald-100 p-1.5 rounded-lg">📄</span>
                            Purchase Details
                        </h2>
                        <p id="purchase-detail-number" class="text-xs text-gray-500">#PO-XXXXXXXX</p>
                    </div>
                    <button onclick="closePurchaseDetailModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <div class="grid grid-cols-2 md:grid-cols-4 gap-3" id="purchase-detail-info">
                    <div class="bg-gray-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">Supplier</p>
                        <p id="pd-supplier" class="font-semibold text-gray-800">-</p>
                    </div>
                    <div class="bg-gray-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">Date</p>
                        <p id="pd-date" class="font-semibold text-gray-800">-</p>
                    </div>
                    <div class="bg-gray-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">Created By</p>
                        <p id="pd-created-by" class="font-semibold text-gray-800">-</p>
                    </div>
                    <div class="bg-gray-50 p-3 rounded-xl text-center">
                        <p class="text-xs text-gray-500">Total Cost</p>
                        <p id="pd-total-cost" class="font-semibold text-emerald-700">KES 0.00</p>
                    </div>
                </div>

                <div class="flex-1 overflow-y-auto border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-xs">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-2">Product</th>
                                <th class="p-2">Barcode</th>
                                <th class="p-2 text-center">Quantity</th>
                                <th class="p-2 text-right">Cost Price</th>
                                <th class="p-2 text-right">Subtotal</th>
                            </tr>
                        </thead>
                        <tbody id="purchase-detail-items">
                            <tr>
                                <td colspan="5" class="p-4 text-center text-gray-500">Loading items...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <div class="flex justify-between items-center border-t border-gray-100 pt-3">
                    <div>
                        <span class="text-xs text-gray-500">Total Items:</span>
                        <span id="pd-total-items" class="font-bold text-gray-800 ml-1">0</span>
                        <span class="text-xs text-gray-500 ml-3">Total Cost:</span>
                        <span id="pd-grand-total" class="font-bold text-emerald-700 ml-1">KES 0.00</span>
                    </div>
                    <button onclick="closePurchaseDetailModal()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">Close</button>
                </div>
            </div>
        </div>

        <!-- CUSTOMER MANAGEMENT MODAL -->
        <div id="customer-modal" class="fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-2xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-indigo-100 p-1.5 rounded-lg">👤</span>
                        <span id="customer-modal-title">Add Customer</span>
                    </h2>
                    <button onclick="closeCustomerModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <form id="customer-form" onsubmit="saveCustomer(event)" class="flex-1 flex flex-col overflow-y-auto">
                    <input type="hidden" id="customer-id" value="0">
                    <div class="space-y-3">
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Full Name *</label><input
                                    type="text" id="customer-name" required placeholder="e.g. John Kamau"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Phone</label><input type="text"
                                    id="customer-phone" placeholder="e.g. 0712345678"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                        </div>
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Email</label><input type="email"
                                    id="customer-email" placeholder="e.g. john@example.com"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">ID Number</label><input
                                    type="text" id="customer-idnumber" placeholder="e.g. 12345678"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Address</label><input type="text"
                                id="customer-address" placeholder="e.g. Nairobi, Kenya"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                        </div>
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Credit Limit
                                    (KES)</label><input type="number" id="customer-credit-limit" value="0" step="0.01"
                                    placeholder="0.00"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Current Balance
                                    (KES)</label><input type="text" id="customer-balance" value="0.00" readonly
                                    class="w-full px-3 py-2 border rounded-lg text-sm bg-gray-50 text-gray-700"></div>
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Notes</label><textarea
                                id="customer-notes" rows="2" placeholder="Additional notes about this customer..."
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"></textarea>
                        </div>
                        <div id="customer-alert" class="hidden p-2 rounded-lg text-sm font-medium"></div>
                    </div>
                    <div class="flex gap-3 pt-4 border-t border-gray-100 mt-4">
                        <button type="button" onclick="closeCustomerModal()"
                            class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">Cancel</button>
                        <button type="submit" id="customer-submit-btn"
                            class="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 rounded-lg transition text-sm">💾
                            Save Customer</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- CUSTOMER LIST MODAL -->
        <div id="customer-list-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-5xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <div>
                        <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-indigo-100 p-1.5 rounded-lg">👥</span>
                            Customers
                        </h2>
                        <p class="text-xs text-gray-500">Manage your customers and their credit</p>
                    </div>
                    <div class="flex items-center gap-2">
                        <button onclick="openAddCustomerModal()"
                            class="bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition">➕
                            Add Customer</button>
                        <button onclick="closeCustomerListModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                    </div>
                </div>

                <div class="flex-1 overflow-y-auto border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-xs">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-2">Name</th>
                                <th class="p-2">Phone</th>
                                <th class="p-2">Email</th>
                                <th class="p-2 text-right">Credit Balance</th>
                                <th class="p-2 text-right">Deposit Balance</th>
                                <th class="p-2 text-right">Credit Limit</th>
                                <th class="p-2 text-center">Actions</th>
                            </tr>
                        </thead>
                        <tbody id="customer-list-body">
                            <tr>
                                <td colspan="7" class="p-4 text-center text-gray-500">Loading customers...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div class="flex justify-end border-t border-gray-100 pt-3">
                    <button onclick="closeCustomerListModal()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">Close</button>
                </div>
            </div>
        </div>

        <!-- CREDIT SALE MODAL -->
        <div id="credit-sale-modal" class="fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-lg rounded-2xl shadow-2xl p-6 space-y-4">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-purple-100 p-1.5 rounded-lg">📝</span>
                        Credit Sale
                    </h2>
                    <button onclick="closeCreditSaleModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <form id="credit-sale-form" onsubmit="submitCreditSale(event)">
                    <div class="space-y-3">
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Customer *</label><select
                                id="credit-customer" required
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                                <option value="">Select Customer</option>
                            </select><button type="button" onclick="openAddCustomerFromCredit()"
                                class="text-xs text-purple-600 hover:text-purple-800 mt-1">➕ Add New Customer</button></div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Total Amount (KES)
                                *</label><input type="number" id="credit-total-amount" required step="0.01" placeholder="0.00"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Amount Paid Today</label><input
                                type="number" id="credit-amount-paid" step="0.01" value="0" placeholder="0.00"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none"
                                oninput="updateCreditBalance()"></div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Remaining Balance *</label><input
                                type="number" id="credit-balance" required step="0.01" placeholder="0.00"
                                class="w-full px-3 py-2 border rounded-lg text-sm bg-gray-50 text-gray-700 focus:ring-2 focus:ring-purple-500 focus:outline-none"
                                readonly></div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Due Date</label><input type="date"
                                id="credit-due-date"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Notes</label><textarea
                                id="credit-notes" rows="2" placeholder="Additional notes..."
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none"></textarea>
                        </div>
                        <div id="credit-alert" class="hidden p-2 rounded-lg text-sm font-medium"></div>
                    </div>
                    <div class="flex gap-3 pt-4 border-t border-gray-100 mt-4">
                        <button type="button" onclick="closeCreditSaleModal()"
                            class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">Cancel</button>
                        <button type="submit" id="credit-submit-btn"
                            class="flex-1 bg-purple-600 hover:bg-purple-700 text-white font-medium py-2 rounded-lg transition text-sm">📝
                            Create Credit Sale</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- DEPOSIT MODAL -->
        <div id="deposit-modal" class="fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-md rounded-2xl shadow-2xl p-6 space-y-4">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-green-100 p-1.5 rounded-lg">💰</span>
                        Customer Deposit
                    </h2>
                    <button onclick="closeDepositModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <form id="deposit-form" onsubmit="submitDeposit(event)">
                    <div class="space-y-3">
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Customer *</label><select
                                id="deposit-customer" required
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-green-500 focus:outline-none">
                                <option value="">Select Customer</option>
                            </select>
                            <div id="deposit-balance-display" class="text-xs text-gray-500 mt-1">Balance: KES 0.00</div>
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Amount (KES) *</label><input
                                type="number" id="deposit-amount" required step="0.01" min="0.01" placeholder="0.00"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-green-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Payment Method</label><select
                                id="deposit-payment-method"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-green-500 focus:outline-none">
                                <option value="cash">Cash</option>
                                <option value="mpesa">M-Pesa</option>
                                <option value="bank">Bank Transfer</option>
                            </select></div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Reference
                                (Optional)</label><input type="text" id="deposit-reference" placeholder="e.g. M-Pesa code"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-green-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Notes</label><textarea
                                id="deposit-notes" rows="2" placeholder="Additional notes..."
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-green-500 focus:outline-none"></textarea>
                        </div>
                        <div id="deposit-alert" class="hidden p-2 rounded-lg text-sm font-medium"></div>
                    </div>
                    <div class="flex gap-3 pt-4 border-t border-gray-100 mt-4">
                        <button type="button" onclick="closeDepositModal()"
                            class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">Cancel</button>
                        <button type="submit" id="deposit-submit-btn"
                            class="flex-1 bg-green-600 hover:bg-green-700 text-white font-medium py-2 rounded-lg transition text-sm">💰
                            Add Deposit</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- CUSTOMER STATEMENT MODAL -->
        <div id="customer-statement-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-5xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <div>
                        <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-indigo-100 p-1.5 rounded-lg">📊</span>
                            Customer Statement
                        </h2>
                        <p class="text-xs text-gray-500">Complete transaction history and balance summary</p>
                    </div>
                    <div class="flex items-center gap-2">
                        <button onclick="printCustomerStatement()"
                            class="bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition flex items-center gap-1">🖨️
                            Print</button>
                        <button onclick="closeCustomerStatement()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                    </div>
                </div>

                <!-- Customer Selector -->
                <div class="flex items-center gap-3 bg-gray-50 p-3 rounded-xl border border-gray-200">
                    <label class="text-sm font-semibold text-gray-700">👤 Customer:</label>
                    <select id="statement-customer-select" onchange="loadCustomerStatement()"
                        class="flex-1 px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none bg-white">
                        <option value="">Select Customer</option>
                    </select>
                    <button onclick="loadCustomerStatement()"
                        class="bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium px-4 py-2 rounded-lg transition">Generate</button>
                </div>

                <!-- Summary Cards -->
                <div class="grid grid-cols-2 md:grid-cols-4 gap-3" id="statement-summary">
                    <div
                        class="bg-gradient-to-br from-blue-50 to-blue-100/50 p-3 rounded-xl text-center border border-blue-200/50">
                        <p class="text-xs font-medium text-blue-600 uppercase tracking-wider">Total Sales</p>
                        <p id="stmt-total-sales" class="text-xl font-bold text-blue-900">KES 0.00</p>
                    </div>
                    <div
                        class="bg-gradient-to-br from-amber-50 to-amber-100/50 p-3 rounded-xl text-center border border-amber-200/50">
                        <p class="text-xs font-medium text-amber-600 uppercase tracking-wider">Total Payments</p>
                        <p id="stmt-total-payments" class="text-xl font-bold text-amber-900">KES 0.00</p>
                    </div>
                    <div
                        class="bg-gradient-to-br from-emerald-50 to-emerald-100/50 p-3 rounded-xl text-center border border-emerald-200/50">
                        <p class="text-xs font-medium text-emerald-600 uppercase tracking-wider">Deposits</p>
                        <p id="stmt-total-deposits" class="text-xl font-bold text-emerald-900">KES 0.00</p>
                    </div>
                    <div
                        class="bg-gradient-to-br from-purple-50 to-purple-100/50 p-3 rounded-xl text-center border border-purple-200/50">
                        <p class="text-xs font-medium text-purple-600 uppercase tracking-wider">Current Balance</p>
                        <p id="stmt-current-balance" class="text-xl font-bold text-purple-900">KES 0.00</p>
                    </div>
                </div>

                <!-- Statement Table -->
                <div class="flex-1 overflow-y-auto border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-xs">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-2">Date</th>
                                <th class="p-2">Transaction</th>
                                <th class="p-2 text-right">Debit (Sales)</th>
                                <th class="p-2 text-right">Credit (Payments)</th>
                                <th class="p-2 text-right">Deposit</th>
                                <th class="p-2 text-right">Balance</th>
                            </tr>
                        </thead>
                        <tbody id="statement-body">
                            <tr>
                                <td colspan="6" class="p-4 text-center text-gray-500">Select a customer and click Generate
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <div class="flex justify-end border-t border-gray-100 pt-3">
                    <button onclick="closeCustomerStatement()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">Close</button>
                </div>
            </div>
        </div>

        <!-- QUICK ADD CUSTOMER MODAL -->
        <div id="quick-customer-modal" class="fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-md rounded-2xl shadow-2xl p-6 space-y-4">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-indigo-100 p-1.5 rounded-lg">👤</span>
                        Quick Add Customer
                    </h2>
                    <button onclick="closeQuickCustomerModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <form id="quick-customer-form" onsubmit="saveQuickCustomer(event)">
                    <div class="space-y-3">
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Phone Number *</label><input
                                type="text" id="quick-customer-phone" required placeholder="e.g. 0712345678"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Full Name *</label><input type="text"
                                id="quick-customer-name" required placeholder="e.g. John Kamau"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Credit Limit</label><input
                                type="number" id="quick-customer-credit-limit" value="2000" step="100"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            <p class="text-xs text-gray-400 mt-1">Default: KES 2,000</p>
                        </div>
                        <div id="quick-customer-alert" class="hidden p-2 rounded-lg text-sm font-medium"></div>
                    </div>
                    <div class="flex gap-3 pt-4 border-t border-gray-100 mt-4">
                        <button type="button" onclick="closeQuickCustomerModal()"
                            class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">Cancel</button>
                        <button type="submit" id="quick-customer-submit-btn"
                            class="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 rounded-lg transition text-sm">💾
                            Add Customer</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- USER MANAGEMENT MODAL -->
        <div id="user-management-modal" class="fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-5xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <div>
                        <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-indigo-100 p-1.5 rounded-lg">👥</span>
                            User Management
                        </h2>
                        <p class="text-xs text-gray-500">Manage users and their roles</p>
                    </div>
                    <div class="flex items-center gap-2">
                        <button onclick="openAddUserModal()"
                            class="bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition">➕
                            Add User</button>
                        <button onclick="closeUserManagement()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                    </div>
                </div>

                <div class="flex-1 overflow-y-auto border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-xs">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-2">Username</th>
                                <th class="p-2">Full Name</th>
                                <th class="p-2">Email</th>
                                <th class="p-2">Role</th>
                                <th class="p-2">Shop</th>
                                <th class="p-2 text-center">Status</th>
                                <th class="p-2 text-center">Actions</th>
                            </tr>
                        </thead>
                        <tbody id="user-list-body">
                            <tr>
                                <td colspan="7" class="p-4 text-center text-gray-500">Loading users...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div class="flex justify-end border-t border-gray-100 pt-3">
                    <button onclick="closeUserManagement()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">Close</button>
                </div>
            </div>
        </div>

        <!-- ADD/EDIT USER MODAL -->
        <div id="user-modal" class="fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-md rounded-2xl shadow-2xl p-6 space-y-4">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-indigo-100 p-1.5 rounded-lg">👤</span>
                        <span id="user-modal-title">Add User</span>
                    </h2>
                    <button onclick="closeUserModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <form id="user-form" onsubmit="saveUser(event)">
                    <input type="hidden" id="user-id" value="0">
                    <div class="space-y-3">
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Username *</label><input type="text"
                                id="user-username" required placeholder="e.g. john_doe"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Full Name *</label><input type="text"
                                id="user-fullname" required placeholder="e.g. John Doe"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Email</label><input type="email"
                                id="user-email" placeholder="e.g. john@example.com"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Password *</label><input
                                type="password" id="user-password" placeholder="Min 6 characters"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            <p class="text-xs text-gray-400 mt-1" id="password-hint">Leave blank to keep current password
                                (when editing)</p>
                        </div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Role</label><select id="user-role"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                                <option value="cashier">Cashier</option>
                                <option value="manager">Manager</option>
                                <option value="admin">Admin</option>
                                <option value="director">Director</option>
                            </select></div>
                        <div><label class="block text-xs font-semibold text-gray-700 mb-1">Shop</label><select id="user-shop"
                                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                                <option value="0">All Shops</option>
                            </select></div>
                        <div id="user-alert" class="hidden p-2 rounded-lg text-sm font-medium"></div>
                    </div>
                    <div class="flex gap-3 pt-4 border-t border-gray-100 mt-4">
                        <button type="button" onclick="closeUserModal()"
                            class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">Cancel</button>
                        <button type="submit" id="user-submit-btn"
                            class="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 rounded-lg transition text-sm">💾
                            Save User</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- COMPANY SETTINGS MODAL -->
        <div id="company-settings-modal" class="fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-3xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-indigo-100 p-1.5 rounded-lg">🏢</span>
                        Company Settings
                    </h2>
                    <button onclick="closeCompanySettings()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <form id="company-settings-form" onsubmit="saveCompanySettings(event)" class="flex-1 overflow-y-auto">
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div class="space-y-3">
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Company Name *</label>
                                <input type="text" id="company-name" required placeholder="e.g. ABC Retail Ltd"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                                <p class="text-xs text-gray-400 mt-0.5">💡 Only admins and directors can change the company
                                    name</p>
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Registration Number</label>
                                <input type="text" id="company-reg-number" placeholder="e.g. PVT-2024-001"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Tax ID (PIN)</label>
                                <input type="text" id="company-tax-id" placeholder="e.g. A12345678X"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Phone</label>
                                <input type="text" id="company-phone" placeholder="e.g. 0712345678"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Email</label>
                                <input type="email" id="company-email" placeholder="e.g. info@abc.co.ke"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Website</label>
                                <input type="text" id="company-website" placeholder="e.g. www.abc.co.ke"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                        </div>
                        <div class="space-y-3">
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Address</label>
                                <textarea id="company-address" rows="3" placeholder="Physical address..."
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"></textarea>
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Currency</label>
                                <select id="company-currency"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                                    <option value="KES">KES - Kenyan Shilling</option>
                                    <option value="USD">USD - US Dollar</option>
                                    <option value="EUR">EUR - Euro</option>
                                    <option value="GBP">GBP - British Pound</option>
                                    <option value="TZS">TZS - Tanzanian Shilling</option>
                                    <option value="UGX">UGX - Ugandan Shilling</option>
                                </select>
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Receipt Footer</label>
                                <textarea id="company-receipt-footer" rows="3" placeholder="Thank you for your business!..."
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"></textarea>
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Company Logo</label>
                                <div class="flex items-center gap-3">
                                    <div id="logo-preview"
                                        class="w-16 h-16 border rounded-lg flex items-center justify-center bg-gray-50">
                                        <span class="text-2xl">🏢</span>
                                    </div>
                                    <input type="file" id="company-logo" accept="image/*"
                                        class="flex-1 text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100">
                                </div>
                                <p class="text-xs text-gray-400 mt-1">Recommended: PNG or JPG, max 2MB</p>
                            </div>
                        </div>
                    </div>

                    <div id="company-alert" class="hidden mt-4 p-2 rounded-lg text-sm font-medium"></div>

                    <div class="flex gap-3 pt-4 border-t border-gray-100 mt-4">
                        <button type="button" onclick="closeCompanySettings()"
                            class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">
                            Cancel
                        </button>
                        <button type="submit" id="company-settings-submit-btn"
                            class="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 rounded-lg transition text-sm">
                            💾 Save Settings
                        </button>
                    </div>
                </form>
            </div>
        </div>

        <!-- COMPANY MANAGEMENT MODAL -->
        <div id="company-management-modal"
            class="fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-5xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <!-- Header -->
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <div>
                        <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-indigo-100 p-1.5 rounded-lg">🏢</span>
                            Company Management
                        </h2>
                        <p class="text-xs text-gray-500">Manage multiple companies (Demo/Live)</p>
                    </div>
                    <!-- ✅ ADD PASSWORD FIELD HERE -->
                    <div class="flex items-center gap-2">
                        <!-- Master Password Field -->
                        <div class="flex items-center gap-1 bg-gray-50 rounded-lg px-2 py-1 border border-gray-200">
                            <span class="text-xs text-gray-500">🔑</span>
                            <input type="password" id="master-password-input" placeholder="Master password"
                                class="text-xs bg-transparent border-0 focus:ring-0 focus:outline-none py-1 px-1 w-28">
                        </div>
                        <button onclick="openAddCompanyModal()"
                            class="bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition">➕
                            Add Company</button>
                        <button onclick="closeCompanyManagement()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                    </div>
                </div>

                <div class="flex-1 overflow-y-auto border rounded-xl relative">
                    <table class="w-full text-left border-collapse text-xs">
                        <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                            <tr>
                                <th class="p-2">Company</th>
                                <th class="p-2">Registration</th>
                                <th class="p-2">Currency</th>
                                <th class="p-2 text-center">Type</th>
                                <th class="p-2 text-center">Status</th>
                                <th class="p-2 text-center">Actions</th>
                            </tr>
                        </thead>
                        <tbody id="company-list-body">
                            <tr>
                                <td colspan="6" class="p-4 text-center text-gray-500">Loading companies...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div class="flex justify-end border-t border-gray-100 pt-3">
                    <button onclick="closeCompanyManagement()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">Close</button>
                </div>
            </div>
        </div>

        <!-- ADD/EDIT COMPANY MODAL -->
        <div id="company-modal" class="fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-2xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-indigo-100 p-1.5 rounded-lg">🏢</span>
                        <span id="company-modal-title">Add Company</span>
                    </h2>
                    <button onclick="closeCompanyModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <form id="company-form" onsubmit="saveCompany(event)" class="flex-1 overflow-y-auto">
                    <input type="hidden" id="company-id" value="0">
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div class="space-y-3">
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Company Name *</label><input
                                    type="text" id="company-name" required placeholder="e.g. ABC Retail Ltd"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Registration
                                    Number</label><input type="text" id="company-reg-number" placeholder="e.g. PVT-2024-001"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Tax ID (PIN)</label><input
                                    type="text" id="company-tax-id" placeholder="e.g. A12345678X"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Phone</label><input type="text"
                                    id="company-phone" placeholder="e.g. 0712345678"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Email</label><input type="email"
                                    id="company-email" placeholder="e.g. info@abc.co.ke"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Website</label><input type="text"
                                    id="company-website" placeholder="e.g. www.abc.co.ke"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                            </div>
                        </div>
                        <div class="space-y-3">
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Address</label><textarea
                                    id="company-address" rows="3" placeholder="Physical address..."
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"></textarea>
                            </div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Currency</label><select
                                    id="company-currency"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                                    <option value="KES">KES - Kenyan Shilling</option>
                                    <option value="USD">USD - US Dollar</option>
                                    <option value="EUR">EUR - Euro</option>
                                    <option value="GBP">GBP - British Pound</option>
                                    <option value="TZS">TZS - Tanzanian Shilling</option>
                                    <option value="UGX">UGX - Ugandan Shilling</option>
                                </select></div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Company Type</label><select
                                    id="company-type"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                                    <option value="0">Live Production</option>
                                    <option value="1">Demo Company</option>
                                </select></div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Status</label><select
                                    id="company-status"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                                    <option value="1">Active</option>
                                    <option value="0">Inactive</option>
                                </select></div>
                            <div><label class="block text-xs font-semibold text-gray-700 mb-1">Receipt
                                    Footer</label><textarea id="company-receipt-footer" rows="2"
                                    placeholder="Thank you for your business!..."
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"></textarea>
                            </div>
                        </div>
                    </div>

                    <div id="company-alert" class="hidden mt-4 p-2 rounded-lg text-sm font-medium"></div>

                    <div class="flex gap-3 pt-4 border-t border-gray-100 mt-4">
                        <button type="button" onclick="closeCompanyModal()"
                            class="flex-1 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium py-2 rounded-lg transition text-sm">Cancel</button>
                        <button type="submit" id="company-submit-btn"
                            class="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2 rounded-lg transition text-sm">💾
                            Save Company</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- COMPANY SETUP WIZARD MODAL -->
        <div id="company-setup-wizard"
            class="fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4 overflow-y-auto">
            <div class="bg-white w-full max-w-3xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <h2 class="text-xl font-bold text-gray-900 flex items-center gap-2">
                        <span class="bg-gradient-to-r from-purple-600 to-indigo-600 text-white p-2 rounded-lg">🏢</span>
                        Company Setup Wizard
                    </h2>
                    <button onclick="closeCompanySetupWizard()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <!-- Progress Steps -->
                <div class="flex items-center justify-center gap-2 py-2">
                    <div class="flex items-center">
                        <div
                            class="w-8 h-8 rounded-full bg-purple-600 text-white flex items-center justify-center text-sm font-bold">
                            1</div>
                        <div class="w-12 h-0.5 bg-purple-600"></div>
                    </div>
                    <div class="flex items-center">
                        <div
                            class="w-8 h-8 rounded-full bg-gray-300 text-gray-600 flex items-center justify-center text-sm font-bold">
                            2</div>
                        <div class="w-12 h-0.5 bg-gray-300"></div>
                    </div>
                    <div class="flex items-center">
                        <div
                            class="w-8 h-8 rounded-full bg-gray-300 text-gray-600 flex items-center justify-center text-sm font-bold">
                            3</div>
                        <div class="w-12 h-0.5 bg-gray-300"></div>
                    </div>
                    <div class="flex items-center">
                        <div
                            class="w-8 h-8 rounded-full bg-gray-300 text-gray-600 flex items-center justify-center text-sm font-bold">
                            4</div>
                    </div>
                </div>

                <!-- Step Labels -->
                <div class="flex justify-between text-xs text-gray-500 px-2 -mt-2">
                    <span>Company Details</span>
                    <span>Shops/Shops</span>
                    <span>Admin User</span>
                    <span>Review & Complete</span>
                </div>

                <!-- Step Content -->
                <div id="setup-wizard-content" class="flex-1 overflow-y-auto">
                    <!-- Step 1: Company Details -->
                    <div id="setup-step-1" class="space-y-3">
                        <h3 class="text-lg font-semibold text-gray-800">Step 1: Company Details</h3>
                        <p class="text-sm text-gray-500">Enter your company information</p>
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Company Name *</label>
                                <input type="text" id="setup-company-name" placeholder="e.g. ABC Retail Ltd"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Registration Number</label>
                                <input type="text" id="setup-company-reg" placeholder="e.g. PVT-2024-001"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Tax ID (PIN)</label>
                                <input type="text" id="setup-company-tax" placeholder="e.g. A12345678X"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Phone</label>
                                <input type="text" id="setup-company-phone" placeholder="e.g. 0712345678"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Email</label>
                                <input type="email" id="setup-company-email" placeholder="e.g. info@abc.co.ke"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Website</label>
                                <input type="text" id="setup-company-website" placeholder="e.g. www.abc.co.ke"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            </div>
                            <div class="md:col-span-2">
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Address</label>
                                <textarea id="setup-company-address" rows="2" placeholder="Physical address..."
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none"></textarea>
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Currency</label>
                                <select id="setup-company-currency"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                                    <option value="KES">KES - Kenyan Shilling</option>
                                    <option value="USD">USD - US Dollar</option>
                                    <option value="EUR">EUR - Euro</option>
                                    <option value="GBP">GBP - British Pound</option>
                                    <option value="TZS">TZS - Tanzanian Shilling</option>
                                    <option value="UGX">UGX - Ugandan Shilling</option>
                                </select>
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Receipt Footer</label>
                                <input type="text" id="setup-company-footer" placeholder="Thank you for your business!"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            </div>
                        </div>
                        <div class="flex justify-end mt-4">
                            <button onclick="setupNextStep(2)"
                                class="bg-purple-600 hover:bg-purple-700 text-white px-6 py-2 rounded-lg text-sm font-medium transition">Next
                                →</button>
                        </div>
                    </div>

                    <!-- Step 2: Shops/Shops -->
                    <div id="setup-step-2" class="space-y-3 hidden">
                        <h3 class="text-lg font-semibold text-gray-800">Step 2: Add Shops/Shops</h3>
                        <p class="text-sm text-gray-500">Add your store locations</p>
                        <div class="bg-blue-50 border border-blue-200 rounded-lg p-3 text-sm text-blue-700">
                            💡 You can add multiple shops. At least one branch is required.
                        </div>
                        <div id="setup-shops-list" class="space-y-2 max-h-48 overflow-y-auto">
                            <p class="text-sm text-gray-400 text-center py-4">No shops added yet</p>
                        </div>
                        <div class="grid grid-cols-1 md:grid-cols-3 gap-2">
                            <input type="text" id="setup-shop-name" placeholder="Shop name *"
                                class="px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            <input type="text" id="setup-shop-location" placeholder="Location"
                                class="px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            <button onclick="addSetupShop()"
                                class="bg-emerald-600 hover:bg-emerald-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition">➕
                                Add Shop</button>
                        </div>
                        <div class="flex justify-between mt-4">
                            <button onclick="setupPrevStep(1)"
                                class="bg-gray-200 hover:bg-gray-300 text-gray-700 px-6 py-2 rounded-lg text-sm font-medium transition">←
                                Back</button>
                            <button onclick="setupNextStep(3)"
                                class="bg-purple-600 hover:bg-purple-700 text-white px-6 py-2 rounded-lg text-sm font-medium transition">Next
                                →</button>
                        </div>
                    </div>

                    <!-- Step 3: Admin User -->
                    <div id="setup-step-3" class="space-y-3 hidden">
                        <h3 class="text-lg font-semibold text-gray-800">Step 3: Create Admin User</h3>
                        <p class="text-sm text-gray-500">Create the first admin user for this company</p>
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Username *</label>
                                <input type="text" id="setup-admin-username" placeholder="e.g. admin"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Full Name *</label>
                                <input type="text" id="setup-admin-name" placeholder="e.g. John Doe"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Email</label>
                                <input type="email" id="setup-admin-email" placeholder="e.g. admin@abc.co.ke"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            </div>
                            <div>
                                <label class="block text-xs font-semibold text-gray-700 mb-1">Password *</label>
                                <input type="password" id="setup-admin-password" placeholder="Min 6 characters"
                                    class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            </div>
                        </div>
                        <div class="flex justify-between mt-4">
                            <button onclick="setupPrevStep(2)"
                                class="bg-gray-200 hover:bg-gray-300 text-gray-700 px-6 py-2 rounded-lg text-sm font-medium transition">←
                                Back</button>
                            <button onclick="setupNextStep(4)"
                                class="bg-purple-600 hover:bg-purple-700 text-white px-6 py-2 rounded-lg text-sm font-medium transition">Next
                                →</button>
                        </div>
                    </div>

                    <!-- Step 4: Review & Complete -->
                    <div id="setup-step-4" class="space-y-3 hidden">
                        <h3 class="text-lg font-semibold text-gray-800">Step 4: Review & Complete</h3>
                        <p class="text-sm text-gray-500">Review the information before creating your company</p>
                        <div id="setup-review-content"
                            class="bg-gray-50 rounded-lg p-4 text-sm space-y-1 max-h-48 overflow-y-auto">
                            <p class="text-gray-500">Loading preview...</p>
                        </div>
                        <div id="setup-wizard-alert" class="hidden p-3 rounded-lg text-sm font-medium"></div>
                        <div class="flex justify-between mt-4">
                            <button onclick="setupPrevStep(3)"
                                class="bg-gray-200 hover:bg-gray-300 text-gray-700 px-6 py-2 rounded-lg text-sm font-medium transition">←
                                Back</button>
                            <button onclick="completeSetupWizard()"
                                class="bg-emerald-600 hover:bg-emerald-700 text-white px-6 py-2 rounded-lg text-sm font-medium transition flex items-center gap-2">
                                🚀 Create Company
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- IMPORT MODAL -->
        <div id="import-modal" class="fixed inset-0 bg-black/60 hidden z-[9999] flex items-center justify-center p-4">
            <div class="bg-white w-full max-w-5xl rounded-2xl shadow-2xl p-6 space-y-4 max-h-[90vh] flex flex-col">

                <!-- Header -->
                <div class="flex justify-between items-center border-b border-gray-100 pb-3">
                    <div>
                        <h2 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                            <span class="bg-emerald-100 p-1.5 rounded-lg">📥</span>
                            Import Products (QuickBooks)
                        </h2>
                        <p class="text-xs text-gray-500">Preview and import products from Excel</p>
                    </div>
                    <button onclick="closeImportModal()" class="text-gray-400 hover:text-gray-600 text-xl">✕</button>
                </div>

                <!-- ✅ PROGRESS BAR - MOVED TO TOP -->
                <div id="import-progress" class="hidden">
                    <div class="bg-blue-50 border border-blue-200 rounded-lg p-3">
                        <div class="flex justify-between items-center mb-1">
                            <span id="import-status" class="text-sm font-medium text-blue-700">Preparing import...</span>
                            <span id="import-percentage" class="text-sm font-bold text-blue-700">0%</span>
                        </div>
                        <div class="w-full bg-blue-200 rounded-full h-3 overflow-hidden">
                            <div id="import-progress-bar"
                                class="bg-gradient-to-r from-blue-500 to-emerald-500 h-3 rounded-full transition-all duration-300"
                                style="width: 0%"></div>
                        </div>
                        <div id="import-details" class="text-xs text-gray-500 mt-1">
                            <span id="imported-count">0</span> / <span id="total-count">0</span> products processed
                        </div>
                    </div>
                </div>

                <!-- Step 1: Upload -->
                <div id="import-step-1" class="space-y-4 flex-1 overflow-y-auto">
                    <div class="bg-blue-50 border border-blue-200 rounded-lg p-3 text-sm text-blue-700">
                        <p class="font-semibold">📋 QuickBooks Format Requirements:</p>
                        <ul class="list-disc list-inside text-xs mt-1 space-y-0.5">
                            <li>Required: <strong>Item Name</strong>, <strong>Average Unit Cost</strong>, <strong>Regular
                                    Price</strong></li>
                            <li>Barcode: <strong>ALU</strong> (primary), <strong>UPC</strong> (secondary), or <strong>Item
                                    Number</strong> (fallback)</li>
                            <li>Optional: Qty 1, Department Name, Vendor Name</li>
                        </ul>
                    </div>

                    <div>
                        <label class="block text-xs font-semibold text-gray-700 mb-1">Select Shop (for stock)</label>
                        <select id="import-shop-select"
                            class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none">
                            <option value="0">No shop stock (just products)</option>
                        </select>
                    </div>

                    <div>
                        <label class="block text-xs font-semibold text-gray-700 mb-1">Excel File *</label>
                        <input type="file" id="import-file" accept=".xlsx,.xls"
                            class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-emerald-500 focus:outline-none">
                        <p class="text-xs text-gray-400 mt-1">Supported: .xlsx, .xls (Max 10MB)</p>
                    </div>

                    <button onclick="openBarcodeUpdateModal()"
                        class="bg-blue-600 hover:bg-blue-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition flex items-center gap-1">
                        🔄 Update Barcodes
                    </button>
                    <button onclick="previewImport()" id="preview-btn"
                        class="w-full bg-emerald-600 hover:bg-emerald-700 text-white font-medium py-2 rounded-lg transition text-sm">
                        🔍 Preview Products
                    </button>
                </div>

                <!-- Step 2: Preview -->
                <div id="import-step-2" class="hidden space-y-4 flex-1 flex flex-col overflow-y-auto">
                    <div class="flex justify-between items-center">
                        <div>
                            <p class="text-sm font-semibold text-gray-700">Preview Results</p>
                            <p id="preview-summary" class="text-xs text-gray-500"></p>
                        </div>
                        <div class="flex gap-2">
                            <button onclick="goBackToUpload()"
                                class="bg-gray-200 hover:bg-gray-300 text-gray-700 text-xs font-medium px-3 py-1.5 rounded-lg transition">←
                                Back</button>
                            <button onclick="confirmImport()" id="confirm-import-btn"
                                class="bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-medium px-3 py-1.5 rounded-lg transition">
                                📤 Import All
                            </button>
                        </div>
                    </div>

                    <div id="preview-categories" class="flex flex-wrap gap-1"></div>

                    <div class="flex-1 overflow-y-auto border rounded-xl relative">
                        <table class="w-full text-left border-collapse text-xs">
                            <thead class="bg-gray-100 sticky top-0 border-b text-xs uppercase text-gray-600 z-10">
                                <tr>
                                    <th class="p-2">Row</th>
                                    <th class="p-2">Product Name</th>
                                    <th class="p-2">Barcode</th>
                                    <th class="p-2">Category</th>
                                    <th class="p-2 text-right">Cost</th>
                                    <th class="p-2 text-right">Retail</th>
                                    <th class="p-2 text-center">Stock</th>
                                    <th class="p-2 text-center">Status</th>
                                </tr>
                            </thead>
                            <tbody id="preview-table-body">
                                <tr>
                                    <td colspan="8" class="p-4 text-center text-gray-500">Loading preview...</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>

                <!-- Result -->
                <div id="import-result" class="hidden p-3 rounded-lg text-sm max-h-48 overflow-y-auto"></div>

                <!-- Footer -->
                <div class="flex justify-end border-t border-gray-100 pt-3">
                    <button onclick="closeImportModal()"
                        class="bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium px-5 py-2 rounded-xl transition text-sm">Close</button>
                </div>
            </div>
        </div>

        <div id="payment-modal" class="fixed inset-0 bg-black bg-opacity-50 hidden flex items-center justify-center p-4 z-50">
            <div class="bg-white rounded-2xl max-w-md w-full p-6 shadow-2xl space-y-6">
                <div class="flex justify-between items-center border-b pb-3">
                    <h3 class="text-xl font-bold text-gray-800">Payment Breakdown</h3>
                    <button onclick="closePaymentModal()"
                        class="text-gray-400 hover:text-gray-600 text-2xl font-bold">&times;</button>
                </div>
                <div class="bg-purple-50 p-4 rounded-xl flex justify-between items-center">
                    <span class="text-sm font-semibold text-purple-900">Total Payable:</span>
                    <span id="modal-total" class="text-xl font-extrabold text-purple-900">KES 0.00</span>
                </div>
                <div class="space-y-4">
                    <div><label class="block text-xs font-semibold uppercase text-gray-500 mb-1">💵 Cash Amount
                            (KES)</label><input type="number" id="cash-input" value="0" step="0.01" oninput="calculateSplit()"
                            class="w-full p-3 border rounded-lg text-lg focus:ring-2 focus:ring-emerald-600 focus:outline-none">
                    </div>
                    <div><label class="block text-xs font-semibold uppercase text-gray-500 mb-1">📲 M-Pesa Amount
                            (KES)</label><input type="number" id="mpesa-input" value="0" step="0.01" oninput="calculateSplit()"
                            class="w-full p-3 border rounded-lg text-lg focus:ring-2 focus:ring-emerald-600 focus:outline-none">
                    </div>
                    <div><label class="block text-xs font-semibold uppercase text-gray-500 mb-1">M-Pesa Ref / Code
                            (Optional)</label><input type="text" id="mpesa-code" placeholder="e.g. QX12345678"
                            class="w-full p-3 border rounded-lg font-mono focus:ring-2 focus:ring-emerald-600 focus:outline-none uppercase">
                    </div>
                    <div id="payment-status" class="text-sm font-medium text-center py-2 rounded-lg bg-gray-100 text-gray-600">
                        Enter Cash and/or
                        M-Pesa amounts</div>
                </div>
                <div id="deposit-section" style="display: none;" class="border-t pt-3">
                    <div class="flex justify-between items-center">
                        <div>
                            <label class="block text-xs font-semibold text-gray-700 mb-1">💰 Use Deposit Balance</label>
                            <p id="deposit-balance-display" class="text-xs text-green-600 font-medium">Available: KES 0.00
                            </p>
                        </div>
                        <div class="flex items-center gap-2">
                            <input type="number" id="deposit-input" value="0" step="0.01" min="0"
                                oninput="updateDepositCalculation()"
                                class="w-32 px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:outline-none">
                            <button
                                onclick="document.getElementById('deposit-input').value = Math.min(customerDepositBalance, totalAmount); updateDepositCalculation();"
                                class="text-xs bg-purple-100 hover:bg-purple-200 text-purple-700 px-2 py-1 rounded-lg transition">Max</button>
                        </div>
                    </div>
                    <div id="deposit-warning" style="display: none;"
                        class="mt-2 text-xs text-amber-600 bg-amber-50 p-2 rounded-lg border border-amber-200">⚠️
                        Insufficient deposit. Remaining amount will be charged to credit.</div>
                    <div id="deposit-remaining-display" class="mt-1 text-xs text-gray-500">Remaining to pay: <span
                            id="remaining-amount">KES 0.00</span></div>
                </div>
                <div class="flex gap-3">
                    <button onclick="closePaymentModal()"
                        class="w-1/3 bg-gray-200 hover:bg-gray-300 text-gray-700 font-bold py-3 rounded-lg transition">Cancel</button>
                    <button id="confirm-sale-btn" onclick="submitSale()" disabled
                        class="w-2/3 bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 text-white font-bold py-3 rounded-lg shadow transition">Confirm
                        & Print</button>
                </div>
            </div>
        </div>

        <!-- ============================================ -->
        <!-- RECENT SALES MODAL -->
        <!-- ============================================ -->
        <div id="recent-sales-modal" class="fixed inset-0 bg-black/50 hidden z-50 flex justify-end">
            <div class="bg-white w-full max-w-lg h-full p-6 overflow-y-auto shadow-2xl flex flex-col justify-between">
                <div>
                    <div class="flex justify-between items-center border-b pb-4 mb-4">
                        <h2 class="text-xl font-bold text-gray-800">Recent Transactions Log</h2>
                        <button onclick="closeRecentSalesModal()"
                            class="text-gray-400 hover:text-gray-600 font-bold text-xl">✕</button>
                    </div>
                    <div id="recent-sales-list" class="space-y-3">
                        <p class="text-gray-500 text-center py-6">Loading past sales...</p>
                    </div>
                </div>
                <div class="border-t pt-4">
                    <button onclick="closeRecentSalesModal()"
                        class="w-full bg-gray-200 text-gray-800 font-semibold py-2.5 rounded-lg hover:bg-gray-300">Close
                        Log</button>
                </div>
            </div>
        </div>

        <!-- ============================================ -->
        <!-- SHOP SELECTOR MODAL -->
        <!-- ============================================ -->
        <div id="shop-selector-modal" class="fixed inset-0 bg-black/60 hidden z-50 flex items-center justify-center p-4">
            <div class="bg-white rounded-2xl shadow-2xl w-full max-w-md p-6">
                <div class="flex justify-between items-center border-b pb-3 mb-4">
                    <h2 class="text-xl font-bold text-gray-800">🏪 Select Shop</h2>
                    <button onclick="closeShopSelector()"
                        class="text-gray-400 hover:text-gray-600 text-2xl font-bold">&times;</button>
                </div>
                <div id="shop-list" class="space-y-2 max-h-80 overflow-y-auto">
                    <p class="text-gray-500 text-center py-4">Loading shops...</p>
                </div>
                <div class="border-t pt-4 mt-4">
                    <button onclick="closeShopSelector()"
                        class="w-full bg-gray-200 text-gray-800 font-semibold py-2.5 rounded-lg hover:bg-gray-300 transition">Close</button>
                </div>
            </div>
        </div>
    `;

    // ---------------------------------------------------------
    // 2. CORE OPEN / CLOSE
    // ---------------------------------------------------------
    function open(id) {
        const el = document.getElementById(id);
        if (!el) { console.warn('[SpideModals] modal not found:', id); return; }
        el.classList.remove('hidden');
        if (!el.classList.contains('flex')) el.classList.add('flex');
    }

    function close(id) {
        const el = document.getElementById(id);
        if (el) el.classList.add('hidden');
    }

    window.SpideModals = { open, close };

    function callIfExists(name) {
        if (typeof window[name] === 'function') window[name]();
    }

    // ---------------------------------------------------------
    // 3. HELPERS
    // ---------------------------------------------------------
    function getCurrentUserRole() {
        if (typeof window.getCurrentUserRole === 'function') {
            return window.getCurrentUserRole();
        }
        // fallback
        try {
            const token = (document.cookie.match(/spide_token=([^;]+)/) || [])[1];
            if (token) {
                const payload = JSON.parse(atob(token.split('.')[1]));
                return payload.role || 'cashier';
            }
            const u = JSON.parse(localStorage.getItem('spide_user') || '{}');
            return u.role || 'cashier';
        } catch (e) { return 'cashier'; }
    }

    // ---------------------------------------------------------
    // 4. OPEN / CLOSE FUNCTIONS (matches onclick="" in modals.html)
    // ---------------------------------------------------------

    // --- Add Product ---
    window.openAddProductModal = function () {
        open('addProductModal');
        callIfExists('loadCategories');

        const form = document.getElementById('addProductForm');
        const isEdit = form && form.dataset.editId && form.dataset.editId !== '';

        const stockEl = document.getElementById('prodStockQty');
        if (stockEl) {
            const wrapper = stockEl.closest('div');
            if (wrapper) wrapper.style.display = isEdit ? 'none' : '';
        }

        setTimeout(() => {
            const el = document.getElementById('prodSku');
            if (el) el.focus();
        }, 50);
    };
    window.closeAddProductModal = function () {
        close('addProductModal');
        const form = document.getElementById('addProductForm');
        if (form) { form.reset(); delete form.dataset.editId; }
        const alert = document.getElementById('modalAlert');
        if (alert) alert.classList.add('hidden');
    };

    // --- Product Catalog ---
    window.openProductCatalogModal = function () {
        open('product-catalog-modal');
        callIfExists('loadProducts');
    };
    window.closeProductCatalogModal = function () { close('product-catalog-modal'); };

    // --- Low Stock ---
    window.openLowStockModal = function () {
        open('low-stock-modal');
        callIfExists('fetchLowStockData');
    };
    window.closeLowStockModal = function () { close('low-stock-modal'); };

    // --- Inventory Valuation ---
    window.openInventoryValuationModal = function () {
        open('inventory-valuation-modal');
        callIfExists('loadShopsForValuation');
        callIfExists('loadInventoryValuation');
    };
    window.closeInventoryValuationModal = function () { close('inventory-valuation-modal'); };

    // --- Z-Report (with role-based shop filter) ---
    window.openZReportModal = function () {
        open('zreport-modal');

        // Show shop filter only for directors/admins
        const filter = document.getElementById('zreport-shop-filter-container');
        if (filter) {
            const role = getCurrentUserRole();
            if (role === 'director' || role === 'admin') {
                filter.classList.remove('hidden');
                filter.classList.add('flex');
            } else {
                filter.classList.add('hidden');
                filter.classList.remove('flex');
            }
        }

        const d = document.getElementById('zreport-date');
        if (d && !d.value) d.value = new Date().toISOString().split('T')[0];

        callIfExists('loadZReportShops');
        callIfExists('loadZReport');
    };
    window.closeZReportModal = function () { close('zreport-modal'); };

    // --- Product Sales ---
    window.openProductSalesModal = function () {
        open('product-sales-modal');

        // Default dates: today
        const today = new Date().toISOString().split('T')[0];
        const startEl = document.getElementById('productsales-start-date');
        const endEl = document.getElementById('productsales-end-date');
        if (startEl && !startEl.value) startEl.value = today;
        if (endEl && !endEl.value) endEl.value = today;

        // Show filters only for directors/admins
        const role = typeof window.getCurrentUserRole === 'function' ? window.getCurrentUserRole() : 'cashier';
        const isDirector = (role === 'director' || role === 'admin');

        const shopContainer = document.getElementById('productsales-shop-filter-container');
        const catContainer = document.getElementById('productsales-category-filter-container');
        const prodContainer = document.getElementById('productsales-product-filter-container');

        [shopContainer, catContainer, prodContainer].forEach(el => {
            if (!el) return;
            if (isDirector) {
                el.classList.remove('hidden');
                el.classList.add('flex');
            } else {
                el.classList.add('hidden');
                el.classList.remove('flex');
            }
        });

        // Load filter options in parallel, then load the report
        const jobs = [];
        if (isDirector) {
            if (typeof window.loadProductSalesShops === 'function') jobs.push(window.loadProductSalesShops());
            if (typeof window.loadProductSalesCategories === 'function') jobs.push(window.loadProductSalesCategories());
        }
        Promise.all(jobs).then(() => {
            if (typeof window.loadProductSalesProducts === 'function') {
                return window.loadProductSalesProducts();
            }
        }).then(() => {
            callIfExists('loadProductSalesReport');
        });
    };
    window.closeProductSalesModal = function () { close('product-sales-modal'); };

    // --- Expense ---
    window.openExpenseModal = function () {
        open('expense-modal');
        const d = document.getElementById('expense-date');
        if (d && !d.value) d.value = new Date().toISOString().split('T')[0];
    };
    window.closeExpenseModal = function () { close('expense-modal'); };

    // --- Expense Report ---
    window.openExpenseReportModal = function () {
        open('expense-report-modal');
        callIfExists('loadExpenseCategories');
        callIfExists('loadExpenseReport');
    };
    window.closeExpenseReportModal = function () { close('expense-report-modal'); };

    // --- Transfer ---
    window.openTransferModal = function () {
        open('transfer-modal');
        callIfExists('loadTransferShops');
    };
    window.closeTransferModal = function () { close('transfer-modal'); };

    window.openTransferHistoryModal = function () {
        open('transfer-history-modal');
        callIfExists('loadTransferHistory');
    };
    window.closeTransferHistoryModal = function () { close('transfer-history-modal'); };

    window.closeTransferDetailModal = function () { close('transfer-detail-modal'); };

    // --- Purchase ---
    window.openPurchaseModal = function () {
        open('purchase-modal');
        callIfExists('loadPurchaseSuppliers');
    };
    window.closePurchaseModal = function () { close('purchase-modal'); };

    window.openPurchaseReportModal = function () {
        open('purchase-report-modal');
        callIfExists('loadPurchaseReport');
    };
    window.closePurchaseReportModal = function () { close('purchase-report-modal'); };

    window.closePurchaseDetailModal = function () { close('purchase-detail-modal'); };

    // --- Supplier ---
    window.openSupplierModal = function () {
        open('supplier-modal');
        const f = document.getElementById('supplier-form');
        if (f) f.reset();
    };
    window.closeSupplierModal = function () { close('supplier-modal'); };

    // --- Customers ---
    window.openCustomerListModal = function () {
        open('customer-list-modal');
        callIfExists('loadCustomers');
    };
    window.closeCustomerListModal = function () { close('customer-list-modal'); };

    window.openCustomerStatement = function () {
        open('customer-statement-modal');
        callIfExists('loadStatementCustomers');
    };
    window.closeCustomerStatement = function () { close('customer-statement-modal'); };

    window.openAddCustomerModal = function () { open('customer-modal'); };
    window.closeCustomerModal = function () { close('customer-modal'); };

    window.openCreditSaleModal = function () { open('credit-sale-modal'); };
    window.closeCreditSaleModal = function () { close('credit-sale-modal'); };

    window.openDepositModal = function () {
        open('deposit-modal');
        callIfExists('loadDepositCustomers');
    };
    window.closeDepositModal = function () { close('deposit-modal'); };

    window.openAddCustomerQuick = function () {
        open('quick-customer-modal');
        setTimeout(() => {
            const el = document.getElementById('quick-customer-phone');
            if (el) el.focus();
        }, 50);
    };
    window.closeQuickCustomerModal = function () { close('quick-customer-modal'); };

    // --- Users ---
    window.openUserManagement = function () {
        open('user-management-modal');
        callIfExists('loadUsers');
        callIfExists('loadUserShops');
    };
    window.closeUserManagement = function () { close('user-management-modal'); };

    window.openAddUserModal = function () {
        const t = document.getElementById('user-modal-title'); if (t) t.textContent = 'Add User';
        const id = document.getElementById('user-id'); if (id) id.value = '0';
        const f = document.getElementById('user-form'); if (f) f.reset();
        const pw = document.getElementById('user-password'); if (pw) pw.required = true;
        const hint = document.getElementById('password-hint');
        if (hint) hint.textContent = 'Enter a password (min 6 characters)';
        const alert = document.getElementById('user-alert');
        if (alert) alert.classList.add('hidden');
        open('user-modal');
    };
    window.closeUserModal = function () { close('user-modal'); };

    // --- Company Settings ---
    window.openCompanySettings = function () {
        open('company-settings-modal');
        callIfExists('loadCompanySettings');
    };
    window.closeCompanySettings = function () { close('company-settings-modal'); };

    // --- Company Management ---
    window.openCompanyManagement = function () {
        open('company-management-modal');
        callIfExists('loadCompanies');
    };
    window.closeCompanyManagement = function () { close('company-management-modal'); };

    window.openAddCompanyModal = function () {
        const t = document.getElementById('company-modal-title'); if (t) t.textContent = 'Add Company';
        const id = document.getElementById('company-id'); if (id) id.value = '0';
        const f = document.getElementById('company-form'); if (f) f.reset();
        open('company-modal');
    };
    window.closeCompanyModal = function () { close('company-modal'); };

    // --- Company Setup Wizard ---
    window.openCompanySetupWizard = function () {
        open('company-setup-wizard');
        if (typeof window.showSetupStep === 'function') window.showSetupStep(1);
    };
    window.closeCompanySetupWizard = function () { close('company-setup-wizard'); };

    // --- Import ---
    window.openImportModal = function () {
        open('import-modal');
        callIfExists('loadImportShops');
        // reset to step 1
        const s1 = document.getElementById('import-step-1');
        const s2 = document.getElementById('import-step-2');
        if (s1) s1.classList.remove('hidden');
        if (s2) s2.classList.add('hidden');
        const prog = document.getElementById('import-progress');
        if (prog) prog.classList.add('hidden');
        const res = document.getElementById('import-result');
        if (res) res.classList.add('hidden');
    };
    window.closeImportModal = function () { close('import-modal'); };

    // --- Recent Sales ---
    window.openRecentSalesModal = function () {
        open('recent-sales-modal');
        callIfExists('loadRecentSales');
    };
    window.closeRecentSalesModal = function () { close('recent-sales-modal'); };

    // --- Shop Selector ---
    window.openShopManagement = function () {
        open('shop-selector-modal');
        callIfExists('renderShopList');
    };
    window.closeShopSelector = function () { close('shop-selector-modal'); };

    // ---------------------------------------------------------
    // 5. MOUNT
    // ---------------------------------------------------------
    function mount() {
        if (document.getElementById('spide-modals-root')) return;
        const root = document.createElement('div');
        root.id = 'spide-modals-root';
        root.innerHTML = MODAL_HTML;
        document.body.appendChild(root);
        const count = root.querySelectorAll('[id]').length;
        console.log('[SpideModals] mounted', count, 'elements');
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', mount);
    } else {
        mount();
    }
})();