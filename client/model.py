import pandas as pd
import random
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split, cross_val_score
from sklearn.preprocessing import StandardScaler
from sklearn.metrics import accuracy_score
import joblib

# Function to generate synthetic data for latency, load, and success rate
def generate_data(n=1000):
    data = {
        'latency_3000': [random.randint(5, 50) for _ in range(n)],
        'latency_4000': [random.randint(5, 50) for _ in range(n)],
        'load_3000': [random.randint(1, 10) for _ in range(n)],
        'load_4000': [random.randint(1, 10) for _ in range(n)],
        'success_rate_3000': [round(random.uniform(0.85, 1.0), 2) for _ in range(n)],
        'success_rate_4000': [round(random.uniform(0.85, 1.0), 2) for _ in range(n)],
    }
    
    # Convert to DataFrame
    df = pd.DataFrame(data)

    # Add derived features (differences)
    df['latency_diff'] = df['latency_3000'] - df['latency_4000']
    df['load_diff'] = df['load_3000'] - df['load_4000']
    df['success_rate_diff'] = df['success_rate_3000'] - df['success_rate_4000']

    # Label based on the node with the lower latency
    df['best_node'] = df.apply(lambda row: 0 if row['latency_3000'] < row['latency_4000'] else 1, axis=1)

    return df

# Generate a larger dataset
df = generate_data(n=10000)  # Larger dataset with 10,000 entries

# Features and labels
X = df[['latency_3000', 'latency_4000', 'load_3000', 'load_4000', 'success_rate_3000', 'success_rate_4000', 'latency_diff', 'load_diff', 'success_rate_diff']]
y = df['best_node']

# Split data into training and test sets
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

# Normalize the features
scaler = StandardScaler()
X_train_scaled = scaler.fit_transform(X_train)
X_test_scaled = scaler.transform(X_test)

# Train a Random Forest Classifier
clf = RandomForestClassifier(n_estimators=100, random_state=42)
clf.fit(X_train_scaled, y_train)

# Evaluate the model with cross-validation
cv_scores = cross_val_score(clf, X_train_scaled, y_train, cv=5)
print(f"Cross-validation scores: {cv_scores}")
print(f"Mean cross-validation score: {cv_scores.mean()}")

# Predict on the test set
y_pred = clf.predict(X_test_scaled)
accuracy = accuracy_score(y_test, y_pred)
print(f"Accuracy: {accuracy}")

# Save the model and scaler
joblib.dump(clf, 'best_node_model.joblib')
joblib.dump(scaler, 'scaler.joblib')
