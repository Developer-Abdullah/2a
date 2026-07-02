import SwiftUI

struct UDIDEnrollmentView: View {
    @StateObject private var viewModel = EnrollmentViewModel()
    @Environment(\.scenePhase) private var scenePhase
    
    // Callback to navigate the user to the Activation Screen upon success
    var onEnrollmentSuccess: ((String) -> Void)?
    
    var body: some View {
        VStack(spacing: 32) {
            headerView
            
            VStack(alignment: .leading, spacing: 24) {
                stepOneView
                stepTwoView
                stepThreeView
            }
            .padding(.horizontal, 24)
            
            Spacer()
            
            actionButton
                .padding(.horizontal, 24)
                .padding(.bottom, 40)
        }
        .background(Color(.systemGroupedBackground).ignoresSafeArea())
        // CRITICAL iOS BEHAVIOR: React to app foregrounding to trigger polling
        .onChange(of: scenePhase) { newPhase in
            if newPhase == .active {
                viewModel.resumePollingIfNeeded()
            }
        }
        // Observe success state to trigger navigation
        .onChange(of: viewModel.state) { newState in
            if case .success(let deviceId) = newState {
                // Add a slight delay for better UX so the user sees the green checkmark
                DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                    onEnrollmentSuccess?(deviceId)
                }
            }
        }
    }
    
    // MARK: - UI Components
    
    private var headerView: some View {
        VStack(spacing: 12) {
            Image(systemName: "lock.shield.fill")
                .resizable()
                .scaledToFit()
                .frame(width: 60, height: 60)
                .foregroundColor(.accentColor)
                .padding(.top, 40)
            
            Text("تسجيل الجهاز")
                .font(.title2)
                .fontWeight(.bold)
            
            Text("لحماية حسابك، نحتاج إلى توثيق هذا الجهاز في خوادم أبل.")
                .font(.subheadline)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
                .padding(.horizontal, 32)
        }
    }
    
    private var stepOneView: some View {
        StepRow(
            icon: "arrow.down.doc.fill",
            title: "الخطوة الأولى",
            subtitle: "تحميل الملف الشخصي",
            isActive: viewModel.state == .idle || viewModel.state == .loadingToken,
            isCompleted: isStepOneCompleted
        )
    }
    
    private var stepTwoView: some View {
        StepRow(
            icon: "gearshape.fill",
            title: "الخطوة الثانية",
            subtitle: "الإعدادات ← عام ← VPN والإدارة ← تثبيت",
            isActive: isStepTwoActive,
            isCompleted: isStepTwoCompleted
        )
    }
    
    private var stepThreeView: some View {
        StepRow(
            icon: isStepThreeCompleted ? "checkmark.circle.fill" : "arrow.triangle.2.circlepath",
            title: "الخطوة الثالثة",
            subtitle: isStepThreeCompleted ? "تم التوثيق بنجاح" : "الانتظار للتحقق من الجهاز...",
            isActive: isStepThreeActive,
            isCompleted: isStepThreeCompleted,
            showSpinner: isStepThreeActive && !isStepThreeCompleted
        )
    }
    
    private var actionButton: some View {
        Button(action: {
            Task {
                await viewModel.downloadProfile()
            }
        }) {
            HStack {
                if viewModel.state == .loadingToken {
                    ProgressView()
                        .progressViewStyle(CircularProgressViewStyle(tint: .white))
                } else {
                    Text("تحميل الملف الآن")
                        .fontWeight(.semibold)
                }
            }
            .frame(maxWidth: .infinity)
            .padding()
            .background(isStepOneCompleted ? Color.gray.opacity(0.3) : Color.accentColor)
            .foregroundColor(isStepOneCompleted ? .primary : .white)
            .cornerRadius(12)
        }
        .disabled(isStepOneCompleted)
    }
    
    // MARK: - State Computed Properties
    
    private var isStepOneCompleted: Bool {
        switch viewModel.state {
        case .downloading, .polling, .success: return true
        default: return false
        }
    }
    
    private var isStepTwoActive: Bool {
        switch viewModel.state {
        case .downloading: return true
        default: return false
        }
    }
    
    private var isStepTwoCompleted: Bool {
        switch viewModel.state {
        case .polling, .success: return true
        default: return false
        }
    }
    
    private var isStepThreeActive: Bool {
        switch viewModel.state {
        case .polling, .success: return true
        default: return false
        }
    }
    
    private var isStepThreeCompleted: Bool {
        if case .success = viewModel.state { return true }
        return false
    }
}

// MARK: - Helper UI View

struct StepRow: View {
    let icon: String
    let title: String
    let subtitle: String
    let isActive: Bool
    let isCompleted: Bool
    var showSpinner: Bool = false
    
    var body: some View {
        HStack(alignment: .top, spacing: 16) {
            ZStack {
                Circle()
                    .fill(isCompleted ? Color.green : (isActive ? Color.accentColor : Color.gray.opacity(0.3)))
                    .frame(width: 44, height: 44)
                
                if showSpinner {
                    ProgressView()
                        .progressViewStyle(CircularProgressViewStyle(tint: .white))
                } else {
                    Image(systemName: isCompleted ? "checkmark" : icon)
                        .foregroundColor(isCompleted || isActive ? .white : .gray)
                        .font(.system(size: 18, weight: .semibold))
                }
            }
            
            VStack(alignment: .leading, spacing: 4) {
                Text(title)
                    .font(.caption)
                    .fontWeight(.bold)
                    .foregroundColor(isActive || isCompleted ? .primary : .secondary)
                
                Text(subtitle)
                    .font(.subheadline)
                    .foregroundColor(isActive || isCompleted ? .primary : .secondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .padding(.top, 4)
            
            Spacer()
        }
        .opacity(isActive || isCompleted ? 1.0 : 0.5)
    }
}